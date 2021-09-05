package app

import (
	"time"

	"github.com/spf13/pflag"
)

type ClientOptions struct {
	Server           string // server address, example example.com:80
	Local            string // local address, for example 3000/:3000/192.168.0.12:8000 are all valid
	KeepAlive        time.Duration
	MaxRetryCount    int
	MaxRetryInterval time.Duration
	Host             string
	Headers          []string
}

func NewClientOptions() *ClientOptions {
	return &ClientOptions{
		KeepAlive:        1 * time.Minute,
		MaxRetryCount:    0,
		MaxRetryInterval: 5 * time.Minute,
	}
}

func (c *ClientOptions) Flags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("client", pflag.ContinueOnError)
	fs.StringVar(&c.Server, "server", c.Server, "Server address, for example, example.com:80")
	fs.StringVar(&c.Local, "local", c.Local, "Local address, 127.0.0.1:8000")
	fs.DurationVar(&c.KeepAlive, "keepalive", c.KeepAlive, "Keepalive duration")
	fs.IntVar(&c.MaxRetryCount, "mex-retry", c.MaxRetryCount, "Maximum retries, 0 means never stop")
	fs.StringVar(&c.Host, "host", c.Host, "Override proxy host")
	fs.DurationVar(&c.MaxRetryInterval, "max-retry-interval", c.MaxRetryInterval, "Maximum duration between two retries")
	fs.StringSliceVar(&c.Headers, "headers", c.Headers, "Custom headers, for example, host=example.com")

	return fs
}
