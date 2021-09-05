package app

import (
	"fmt"
	"net"

	"github.com/spf13/pflag"
	"k8s.io/klog"
)

type KunnelOptions struct {
	Domain     string // top level domain
	Bind       string // server address, default 127.0.0.1
	Port       int    // server port
	TlsKeyFile string
	TlsCrtFile string
}

func NewKunnelOptions() *KunnelOptions {
	return &KunnelOptions{
		Bind: "127.0.0.1",
		Port: 80,
	}
}

func (k *KunnelOptions) Flags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("kunnel", pflag.ContinueOnError)
	flags.StringVar(&k.Domain, "domain", k.Domain, "Tunnel top level domain name. *.[domain] MUST resolve to server address.")
	flags.StringVar(&k.Bind, "bind", k.Bind, "Server bind address, default 127.0.0.1")
	flags.IntVar(&k.Port, "port", k.Port, "Server port, default 80.")
	flags.StringVar(&k.TlsCrtFile, "tls-crt-file", k.TlsCrtFile, "Tls certificate crt file")
	flags.StringVar(&k.TlsKeyFile, "tls-key-file", k.TlsKeyFile, "Tls certificate key file")
	return flags
}

func (k *KunnelOptions) Validate() error {
	if err := net.ParseIP(k.Bind); err != nil {
		return fmt.Errorf("invalid bind address %s, %v", k.Bind, err)
	}

	if k.Port <= 0 || k.Port > 65535 {
		return fmt.Errorf("invalid port number %d, must be in the range [0, 65535]", k.Port)
	}

	return nil
}

func (k *KunnelOptions) Print() {
	klog.Infof("--domain=%s", k.Domain)
	klog.Infof("--bind=%s", k.Bind)
	klog.Infof("--port=%d", k.Port)
	klog.Infof("--tls-crt-file=%s", k.TlsCrtFile)
	klog.Infof("--tls-key-file=%s", k.TlsKeyFile)
}
