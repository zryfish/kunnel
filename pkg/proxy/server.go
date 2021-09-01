package proxy

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	client "github.com/zryfish/kunnel/pkg/agent"
	"github.com/zryfish/kunnel/pkg/utils"
	"github.com/zryfish/kunnel/pkg/version"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Options struct {
	Host   string
	Port   int
	Domain string
}

type Server struct {
	httpServer *HttpServer
	sshConfig  *ssh.ServerConfig
	host       string
	port       int
	domain     string
	sessions   map[string]*HttpProxy
}

func NewServer(options *Options) (*Server, error) {
	s := &Server{
		httpServer: NewHttpServer(),
		host:       options.Host,
		port:       options.Port,
		domain:     options.Domain,
		sessions:   make(map[string]*HttpProxy),
	}

	key, _ := generateKey()
	private, err := ssh.ParsePrivateKey(key)
	if err != nil {
		klog.Fatalf("Failed to parse ssh key %v", err)
	}

	s.sshConfig = &ssh.ServerConfig{
		ServerVersion:    "SSH-" + version.ProtocolVersion + "-server",
		PasswordCallback: s.authenticate,
	}
	s.sshConfig.AddHostKey(private)

	return s, nil

}

func (s *Server) authenticate(c ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	klog.V(4).Infof("%s is connecting from %s", c.User(), c.RemoteAddr())
	return nil, nil
}

func (s *Server) handleClientHandler(w http.ResponseWriter, r *http.Request) {
	upgrade := strings.ToLower(r.Header.Get("Upgrade"))
	protocol := r.Header.Get("Sec-WebSocket-Protocol")
	if upgrade == "websocket" && strings.HasPrefix(protocol, "kunnel-") {
		if protocol == version.ProtocolVersion {
			s.handleWebsocket(w, r)
			return
		}
		klog.V(4).Infof("Ingoring client connection using protocol '%s', expected '%s'", protocol, version.ProtocolVersion)
	}

	switch r.URL.String() {
	case "/health":
		w.Write([]byte("OK\n"))
		return
	case "/version":
		w.Write([]byte(version.BuildVersion))
		return
	default:
		s.handleRequest(w, r)
	}
}

func (s *Server) handleWebsocket(w http.ResponseWriter, req *http.Request) {
	klog.V(4).Info("New connection")
	wsConn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		klog.Error("Failed to upgrade connection", err)
		return
	}

	connection := utils.NewWebSocketConn(wsConn)

	sshConn, chans, reqs, err := ssh.NewServerConn(connection, s.sshConfig)
	if err != nil {
		klog.Error("Failed to handshake with client", err)
		return
	}

	var sreq *ssh.Request
	select {
	case sreq = <-reqs:
	case <-time.After(10 * time.Second):
		sshConn.Close()
		return
	}

	if sreq.Type != "config" {
		s.Reply(sreq, "", errors.New("expecting config request"))
		return
	}

	config := &client.Config{}
	if err := config.Unmarshal(sreq.Payload); err != nil {
		klog.Error("Unable to unmarshal config from client", err)
		return
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
			return utils.NewSshConn(sshConn, fmt.Sprintf("%s:%d", config.LocalHost, config.LocalPort)), nil
		},
	}

	proxy := NewHttpProxy(config.Name, config.LocalHost, config.LocalPort, transport)

	subDomain := s.generateSubDomain() + s.domain

	s.Reply(sreq, subDomain, nil)
	s.sessions[subDomain] = proxy

	go s.handleSSHRequests(reqs)
	go s.handleSSHChannels(chans)
	sshConn.Wait()
}

func (s *Server) Run(ctx context.Context) error {
	if err := s.Start(s.host, s.port); err != nil {
		return err
	}
	return nil
}

func (s *Server) Reply(sreq *ssh.Request, domain string, err error) {
	if sreq != nil {
		message := &utils.Message{
			Domain: domain,
			Err:    err,
		}
		ok := true
		if err != nil {
			ok = false
		}

		body, err := message.Marshal()
		if err != nil {
			klog.Error(err)
			return
		}
		sreq.Reply(ok, body)
	}
}

func (s *Server) Start(host string, port int) error {
	h := http.Handler(http.HandlerFunc(s.handleClientHandler))
	h = wrap(h)

	return s.httpServer.GoListenAndServe(fmt.Sprintf("%s:%d", host, port), h)
}

func (s *Server) Wait() error {
	return s.httpServer.Wait()
}

func (s *Server) Close() error {
	return s.httpServer.Close()
}

func wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t0 := time.Now()
		next.ServeHTTP(w, req)

		klog.V(4).Infof("Connection to %s lasts for %s", req.Host, time.Since(t0))

	})
}

func (s *Server) handleSSHRequests(reqs <-chan *ssh.Request) {
	for req := range reqs {
		switch req.Type {
		case "ping":
			req.Reply(true, nil)
		default:
			klog.V(4).Info("Unknown request", req)
		}
	}
}

func (s *Server) handleRequest(w http.ResponseWriter, req *http.Request) {
	host := req.Host

	session, ok := s.sessions[host]
	if !ok {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("No upstream found"))
		return
	}

	session.ServeHTTP(w, req)
}

func (s *Server) handleSSHChannels(chans <-chan ssh.NewChannel) {
	for ch := range chans {
		remote := string(ch.ExtraData())
		stream, reqs, err := ch.Accept()
		if err != nil {
			klog.Error("failed to accept stream", err)
			continue
		}
		go ssh.DiscardRequests(reqs)
		go utils.HandleTCPStream(stream, remote)
	}
}

func generateKey() ([]byte, error) {
	r := rand.Reader

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), r)
	if err != nil {
		return nil, err
	}
	b, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal ECDSA private key: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b}), nil
}

func (s *Server) generateSubDomain() string {
	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	letters := []rune("abcdefghijklmnopqrstuvwxyz1234567890")
	r := make([]rune, 10)
	for i := range r {
		r[i] = letters[int(b[i])*len(letters)/256]
	}
	return string(r) + "."
}
