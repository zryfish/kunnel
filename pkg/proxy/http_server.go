package proxy

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"sync"
)

// HttpServer extends net/http with
// graceful shutdowns
type HttpServer struct {
	*http.Server
	listener  net.Listener
	running   chan error
	isRunning bool
	closer    sync.Once
}

func NewHttpServer() *HttpServer {
	return &HttpServer{
		Server:    &http.Server{},
		listener:  nil,
		running:   make(chan error, 1),
		isRunning: false,
		closer:    sync.Once{},
	}
}

func (h *HttpServer) GoListenAndServeTls(addr string, handler http.Handler, tlsConfig *tls.Config) error {
	var l net.Listener
	var err error
	if tlsConfig != nil {
		l, err = tls.Listen("tcp", addr, tlsConfig)
	} else {
		l, err = net.Listen("tcp", addr)
	}

	if err != nil {
		return err
	}

	h.isRunning = true
	h.Handler = handler
	h.listener = l
	go func() {
		h.CloseWith(h.Serve(l))
	}()
	return nil

}

func (h *HttpServer) CloseWith(err error) {
	if !h.isRunning {
		return
	}
	h.isRunning = false
	h.running <- err
}

func (h *HttpServer) Close() error {
	h.CloseWith(nil)
	return h.listener.Close()
}

func (h *HttpServer) Wait() error {
	if !h.isRunning {
		return errors.New("server already closed")
	}
	return <-h.running
}
