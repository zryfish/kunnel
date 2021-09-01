package app

import "github.com/spf13/pflag"

type KunnelOptions struct {
	Domain string // top level domain
	Host   string // server address, default 127.0.0.1
	Port   int    // server port
}

func NewKunnelOptions() *KunnelOptions {
	return &KunnelOptions{
		Host: "127.0.0.1",
		Port: 80,
	}
}

func (k *KunnelOptions) Flags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("kunnel", pflag.ContinueOnError)
	flags.StringVar(&k.Domain, "domain", k.Domain, "Tunnel top level domain name. *.[domain] MUST resolve to server address.")
	flags.StringVar(&k.Host, "host", k.Host, "Server address, default 127.0.0.1")
	flags.IntVar(&k.Port, "port", k.Port, "Server port, default 80.")
	return flags
}

func (k *KunnelOptions) Validate() error {
	return nil
}

func (k *KunnelOptions) Print() {

}
