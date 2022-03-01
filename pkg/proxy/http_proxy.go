package proxy

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"k8s.io/klog"

	k8sproxy "k8s.io/apimachinery/pkg/util/proxy"
)

type HttpProxy struct {
	name       string
	port       int
	host       string
	protocol   string
	server     *http.Server
	httpClient *http.Client
	proxyHost  string
	headers    map[string]string
}

func NewHttpProxy(name, host, protocol string, port int, proxyHost string, headers map[string]string, transport *http.Transport) *HttpProxy {
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	return &HttpProxy{
		name:      name,
		host:      host,
		port:      port,
		protocol:  protocol,
		headers:   headers,
		server:    server,
		proxyHost: proxyHost,
		httpClient: &http.Client{
			Transport: transport,
		},
	}
}

func (s *HttpProxy) Start(ctx context.Context) error {
	klog.V(0).Infof("Proxy server %s: starting http proxy on %s, proxy address %s", s.name, s.server.Addr, s.host)
	s.server.Handler = s

	done := make(chan chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			if err := s.server.Shutdown(ctx); err != nil {
				klog.Errorf("proxy %s, err: %v", s.name, err)
				return
			}
		case <-done:
			klog.V(2).Infof("proxy %s shutdown", s.name)
		}
	}()

	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			klog.Errorf("proxy server %s: %v", s.name, err)
			close(done)
		}
	}()

	return nil
}

func (s *HttpProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	u := *req.URL
	u.Host = fmt.Sprintf("%s:%d", s.host, s.port)
	u.Scheme = s.protocol

	httpProxy := k8sproxy.NewUpgradeAwareHandler(&u, s.httpClient.Transport, false, false, s)

	if len(s.proxyHost) != 0 {
		req.Host = s.proxyHost
	}

	for k, v := range s.headers {
		if _, existed := req.Header[k]; existed && strings.ToLower(k) != "host" {
			continue
		}
		req.Header.Add(k, v)
	}

	httpProxy.ServeHTTP(w, req)
}

func (s *HttpProxy) Error(_ http.ResponseWriter, req *http.Request, err error) {
	klog.Errorf("Proxy server %s: proxy %s encountered error %v", s.name, req.URL, err)
}
