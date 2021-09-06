package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/zryfish/kunnel/cmd/server/app"
	"github.com/zryfish/kunnel/pkg/proxy"
	"k8s.io/klog"
)

func main() {

	options := app.NewKunnelOptions()

	rootCommand := &cobra.Command{
		Use:  "kunnel",
		Long: "A tool for tunnel Kubernetes service.",
		RunE: func(cmd *cobra.Command, args []string) error {
			options.Print()
			serverOption := &proxy.Options{
				Host:       options.Bind,
				Port:       options.Port,
				Domain:     options.Domain,
				TlsKeyFile: options.TlsKeyFile,
				TlsCrtFile: options.TlsCrtFile,
			}

			srv, err := proxy.NewServer(serverOption)
			if err != nil {
				return err
			}

			if err := srv.Run(context.Background()); err != nil {
				klog.Fatalf("Failed to create proxy %v", err)
			}

			klog.Infof("Server started listening on %s", fmt.Sprintf("%s:%d", options.Bind, options.Port))

			return srv.Wait()
		},
	}

	fs := rootCommand.Flags()
	fs.AddFlagSet(options.Flags())
	klog.InitFlags(nil)
	fs.AddGoFlagSet(flag.CommandLine)

	if err := rootCommand.Execute(); err != nil {
		log.Fatalln(err)
	}
}
