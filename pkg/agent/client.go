package agent

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jpillora/backoff"
	"github.com/zryfish/kunnel/pkg/utils"
	"github.com/zryfish/kunnel/pkg/version"
	"golang.org/x/crypto/ssh"
	"k8s.io/klog"
)

type Client struct {
	sshConfig        *ssh.ClientConfig
	sshConn          ssh.Conn
	running          bool
	runningCh        chan error
	config           *Config
	keepAlive        time.Duration
	maxRetryCount    int
	maxRetryInterval time.Duration
	server           string
}

func NewClient(config *Config, keepAlive time.Duration, maxRetryCount int, maxRetryInterval time.Duration, server string) *Client {
	client := &Client{
		running:          true,
		runningCh:        make(chan error, 1),
		config:           config,
		keepAlive:        keepAlive,
		maxRetryCount:    maxRetryCount,
		maxRetryInterval: maxRetryInterval,
		server:           server,
	}

	client.sshConfig = &ssh.ClientConfig{
		User:            "",
		Auth:            []ssh.AuthMethod{ssh.Password("")},
		ClientVersion:   "SSH-" + version.ProtocolVersion + "-client",
		HostKeyCallback: client.verifyServer,
		Timeout:         30 * time.Second,
	}

	return client
}

func (c *Client) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := c.Start(ctx); err != nil {
		return err
	}

	return c.Wait()
}

func (c *Client) Start(ctx context.Context) error {
	if c.keepAlive > 0 {
		go c.keepAliveLoop()
	}

	go c.connectionLoop()
	return nil
}

func (c *Client) Wait() error {
	return <-c.runningCh
}

func (c *Client) Close() error {
	c.running = false
	if c.sshConn == nil {
		return nil
	}
	return c.sshConn.Close()
}

func (c *Client) keepAliveLoop() {
	for c.running {
		time.Sleep(c.keepAlive)
		if c.sshConn != nil {
			c.sshConn.SendRequest("ping", true, nil)
		}
	}
}

func (c *Client) connectionLoop() {
	var connectionErr error
	b := &backoff.Backoff{Max: c.maxRetryInterval}
	for c.running {
		if connectionErr != nil {
			attempt := int(b.Attempt())
			maxAttempt := c.maxRetryCount
			d := b.Duration()

			msg := fmt.Sprintf("Connection error: %s", connectionErr)
			if attempt > 0 {
				msg += fmt.Sprintf(" (Attempt: %d", attempt)
				if maxAttempt > 0 {
					msg += fmt.Sprintf("/%d", maxAttempt)
				}
				msg += ")"
			}
			klog.Warning(msg)

			if maxAttempt > 0 && attempt >= maxAttempt {
				break
			}
			klog.Warningf("Retrying in %s...", d)
			connectionErr = nil

			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGHUP)
			select {
			case <-time.After(d):
			case <-sig:
			}
			signal.Stop(sig)
		}

		dialer := websocket.Dialer{
			ReadBufferSize:   1024,
			WriteBufferSize:  1024,
			HandshakeTimeout: 45 * time.Second,
			Subprotocols:     []string{version.ProtocolVersion},
		}

		wsHeaders := http.Header{}

		wsConn, _, err := dialer.Dial(c.server, wsHeaders)
		if err != nil {
			connectionErr = err
			continue
		}

		conn := utils.NewWebSocketConn(wsConn)
		klog.V(4).Info("Handshaking...")
		sshConn, chans, reqs, err := ssh.NewClientConn(conn, "", c.sshConfig)
		if err != nil {
			if strings.Contains(err.Error(), "unable to authenticate") {
				klog.Error("Authentication failed", err)
			} else {
				klog.Error(err)
			}
			break
		}

		conf, _ := c.config.Marshal()
		klog.V(4).Info("Sending config")
		t0 := time.Now()
		_, payload, err := sshConn.SendRequest("config", true, conf)
		if err != nil {
			klog.Error("Config verification failed", err)
			break
		}

		msg := &utils.Message{}
		if err := msg.Unmarshal(payload); err != nil {
			klog.Error("Invalid response from server", err)
			break
		}

		if msg.Err != nil {
			klog.Error(msg.Err.Error())
			break
		}

		if len(msg.Domain) != 0 {
			klog.Infof("Service available at https://%s", msg.Domain)
		}

		klog.V(2).Infof("Connected (Latency %s)", time.Since(t0))
		b.Reset()
		c.sshConn = sshConn
		go ssh.DiscardRequests(reqs)
		go c.connectStreams(chans)

		err = sshConn.Wait()
		c.sshConn = nil
		if err != nil && err != io.EOF {
			connectionErr = err
			continue
		}
		klog.V(2).Info("Disconnected")
	}
	close(c.runningCh)
}

func (c *Client) connectStreams(chans <-chan ssh.NewChannel) {
	for ch := range chans {
		remote := string(ch.ExtraData())
		stream, reqs, err := ch.Accept()
		if err != nil {
			klog.Error("Failed to accept stream", err)
			continue
		}

		go ssh.DiscardRequests(reqs)
		go utils.HandleTCPStream(stream, remote)
	}
}

func (agent *Client) verifyServer(hostname string, remote net.Addr, key ssh.PublicKey) error {
	return nil
}
