package proxy

import (
	"context"
	"fmt"
	"net/http"

	"k8s.io/klog"

	k8sproxy "k8s.io/apimachinery/pkg/util/proxy"
)

type HttpProxy struct {
	name       string
	port       int
	host       string
	server     *http.Server
	httpClient *http.Client
}

func NewHttpProxy(name, host string, port int, transport *http.Transport) *HttpProxy {
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	return &HttpProxy{
		name:   name,
		host:   host,
		port:   port,
		server: server,
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

	httpProxy := k8sproxy.NewUpgradeAwareHandler(&u, s.httpClient.Transport, false, false, s)
	httpProxy.ServeHTTP(w, req)
}

func (s *HttpProxy) Error(_ http.ResponseWriter, req *http.Request, err error) {
	klog.Errorf("Proxy server %s: proxy %s encountered error %v", s.name, req.URL, err)
}
