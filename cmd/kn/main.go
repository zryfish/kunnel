package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/zryfish/kunnel/cmd/kn/app"
	"github.com/zryfish/kunnel/pkg/agent"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	knOptions := app.NewKnOptions()

	knCommand := &cobra.Command{
		Use:  "kubectl-kn",
		Long: "kn is a kubectl plugin to proxy kubernetes service outside the cluster.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(knOptions.Service) == 0 {
				return fmt.Errorf("service not provided")
			}

			if len(knOptions.Namespace) == 0 {
				knOptions.Namespace = "default"
			}

			ctx := signals.SetupSignalHandler()

			config, err := clientcmd.BuildConfigFromFlags("", knOptions.KubeConfig)
			if err != nil {
				return err
			}

			kubeClient := kubernetes.NewForConfigOrDie(config)

			service, err := kubeClient.CoreV1().Services(knOptions.Namespace).Get(ctx, knOptions.Service, metav1.GetOptions{})
			if err != nil {
				return err
			}

			if len(service.Spec.ClusterIP) == 0 {
				return fmt.Errorf("headless service is not supported")
			}

			if knOptions.Port == 0 {
				for _, port := range service.Spec.Ports {
					if port.Protocol == v1.ProtocolTCP {
						knOptions.Port = int(port.Port)
						klog.Warningf("No port specified, will use first port [%d] of service", knOptions.Port)
						break
					}
				}

				if knOptions.Port == 0 {
					return fmt.Errorf("no port sepecified")
				}
			}

			headers := make(map[string]string)
			for _, header := range knOptions.Headers {
				parts := strings.Split(header, "=")
				if len(parts) != 2 {
					continue
				}
				headers[parts[0]] = parts[1]
			}

			if knOptions.Daemon {
				return StartInCluster(kubeClient, ctx, knOptions.Namespace, knOptions.Service, service.Spec.ClusterIP, knOptions.Server, knOptions.Port, knOptions.Host, knOptions.Headers)
			}

			return Start(ctx, service.Spec.ClusterIP, knOptions.Port, knOptions.Host, knOptions.Server, headers)
		},
	}

	fs := knCommand.Flags()
	fs.AddFlagSet(knOptions.Flags())

	if err := knCommand.Execute(); err != nil {
		klog.Fatal(err)
	}
}

func Start(ctx context.Context, localhost string, localport int, host, server string, headers map[string]string) error {
	config := &agent.Config{
		Name:      "dummy",
		LocalHost: localhost,
		LocalPort: localport,
		Host:      host,
		Hedaers:   headers,
	}

	agent := agent.NewClient(config, time.Second*3, 20, time.Minute*5, server)
	if err := agent.Run(); err != nil {
		return err
	}

	return agent.Wait()
}

func StartInCluster(kubeClient kubernetes.Interface, ctx context.Context, namespace, service, server string, localhost string, localport int, host string, headers []string) error {
	deployment := app.NewDeployment(namespace, service, localhost, server, localport, host, headers)

	_, err := kubeClient.AppsV1().Deployments(namespace).Get(ctx, deployment.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) { // no deployment existed, created
			_, err = kubeClient.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
			return err
		}
		return err
	}

	// there is already a deployment existed, override
	_, err = kubeClient.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	return err
}
