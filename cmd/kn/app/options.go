package app

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

type KnOptions struct {
	Namespace  string
	KubeConfig string
	Server     string
	Service    string
	Daemon     bool
	Port       int
	Host       string
	Headers    []string
}

func NewKnOptions() *KnOptions {
	return &KnOptions{
		Namespace: "default",
		Server:    "wss://kunnel.run",
	}
}

func (k *KnOptions) Flags() *pflag.FlagSet {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/root/"
	}

	fs := pflag.NewFlagSet("kn", pflag.ContinueOnError)
	fs.StringVarP(&k.Namespace, "namespace", "n", k.Namespace, "Namespace of the service.")
	fs.BoolVarP(&k.Daemon, "daemon", "d", k.Daemon, "Run as a deployment in the cluster.")
	fs.StringVar(&k.Server, "server", k.Server, "Availabel kunnel server address.")
	fs.StringVar(&k.KubeConfig, "kubeconfig", fmt.Sprintf("%s/.kube/config", homeDir), "Location of the kubeconfig")
	fs.StringVarP(&k.Service, "service", "s", k.Service, "Service name to be proxied, only services with cluster ip are supported.")
	fs.StringVar(&k.Host, "host", k.Host, "Override request host field when proxied to destintation.")
	fs.IntVarP(&k.Port, "port", "p", k.Port, "Service port.")
	fs.StringSliceVar(&k.Headers, "headers", []string{}, "Custom headers to be added, format like key=val")
	return fs
}
