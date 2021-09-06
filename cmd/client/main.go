package main

import (
	"flag"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/zryfish/kunnel/cmd/client/app"
	"github.com/zryfish/kunnel/pkg/agent"
	"github.com/zryfish/kunnel/pkg/utils"
	"k8s.io/klog"
)

func main() {
	clientOptions := app.NewClientOptions()

	command := &cobra.Command{
		Use:  "client",
		Long: "A tool for tunnel local service.",
		RunE: func(cmd *cobra.Command, args []string) error {
			local, err := utils.NewLocal(clientOptions.Local)
			if err != nil {
				return err
			}

			headers := make(map[string]string)
			for _, header := range clientOptions.Headers {
				parts := strings.Split(header, "=")
				if len(parts) != 2 {
					continue
				}
				headers[parts[0]] = parts[1]
			}

			config := &agent.Config{
				Name:      "dummy",
				LocalHost: local.LocalHost,
				LocalPort: local.LocalPort,
				Host:      clientOptions.Host,
				Hedaers:   headers,
			}

			agent := agent.NewClient(config, time.Second*3, 20, time.Minute*5, clientOptions.Server)
			if err := agent.Run(); err != nil {
				return err
			}

			return agent.Wait()
		},
	}

	fs := command.Flags()
	fs.AddFlagSet(clientOptions.Flags())
	klog.InitFlags(nil)
	fs.AddGoFlagSet(flag.CommandLine)

	if err := command.Execute(); err != nil {
		klog.Fatal("Failed to run client", err)
	}
}
