package app

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
)

type KnOptions struct {
	Server           string
	Port             int
	Host             string
	Headers          []string
	Local            string // local address, for example 3000/:3000/192.168.0.12:8000 are all valid
	Protocol         string
	KeepAlive        time.Duration
	MaxRetryCount    int
	MaxRetryInterval time.Duration

	Namespace  string
	KubeConfig string
	Service    string
	Daemon     bool
}

func NewKnOptions() *KnOptions {
	return &KnOptions{
		Server:           "wss://kunnel.run",
		MaxRetryInterval: 5 * time.Minute,
		MaxRetryCount:    0,
		KeepAlive:        1 * time.Minute,
		Protocol:         "http",
	}
}

func (k *KnOptions) Flags() *pflag.FlagSet {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/root/"
	}

	fs := pflag.NewFlagSet("kn", pflag.ContinueOnError)
	fs.StringVarP(&k.Namespace, "namespace", "n", k.Namespace, "[Kubernetes Only] Namespace of the service.")
	fs.BoolVarP(&k.Daemon, "daemon", "d", k.Daemon, "Run as a deployment in the cluster.")
	fs.StringVar(&k.Server, "server", k.Server, "Available kunnel server address.")
	fs.StringVar(&k.Protocol, "protocol", k.Protocol, "Proxied service's protocol, only http and https are supported.")
	fs.StringVar(&k.KubeConfig, "kubeconfig", fmt.Sprintf("%s/.kube/config", homeDir), "[Kubernetes Only] Location of the kubeconfig")
	fs.StringVarP(&k.Service, "service", "s", k.Service, "[Kubernetes Only] Service name to be proxied, only services with cluster ip are supported.")
	fs.StringVar(&k.Host, "host", k.Host, "Override request host field when proxied to destintation.")
	fs.IntVarP(&k.Port, "port", "p", k.Port, "[Kubernetes Only] Service port.")
	fs.StringSliceVar(&k.Headers, "headers", []string{}, "Custom headers to be added, format like key=val.")
	fs.StringVar(&k.Local, "local", k.Local, "Local address, 127.0.0.1:8000")
	fs.DurationVar(&k.KeepAlive, "keepalive", k.KeepAlive, "Keepalive duration.")
	fs.IntVar(&k.MaxRetryCount, "mex-retry", k.MaxRetryCount, "Maximum retries, 0 means never stop.")
	fs.DurationVar(&k.MaxRetryInterval, "max-retry-interval", k.MaxRetryInterval, "Maximum duration between two retries.")
	return fs
}
