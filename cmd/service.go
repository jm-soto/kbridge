package cmd

import (
	"fmt"

	"kbridge/pkg/k8s"
	"kbridge/pkg/ui"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

var (
	svcNamespace  string
	svcLocalPort  int
	svcRemotePort int
)

func newServiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service [service-name]",
		Short: "Port forward to a service",
		Long: `Port forward to a service by name. If no name is provided, 
it will list all services in the namespace.

Examples:
  # List all services in the namespace
  kbridge service -n default

  # Port forward to a specific service
  kbridge service nginx -n default`,
		Args: cobra.MaximumNArgs(1),
		RunE: runServicePortForward,
	}

	cmd.Flags().StringVarP(&svcNamespace, "namespace", "n", "default", "The namespace where the service is located")
	cmd.Flags().IntVarP(&svcLocalPort, "local-port", "L", 0, "Local port to forward from (random if not specified)")
	cmd.Flags().IntVarP(&svcRemotePort, "port", "p", 0, "Remote port to forward to (service port if not specified)")

	return cmd
}

func selectServiceInteractive(client *k8s.Client, namespace string) (*corev1.Service, error) {
	services, err := client.ListServices(namespace, []string{})
	if err != nil {
		ui.PrintError("Error listing services: %v", err)
		return nil, err
	}

	if len(services) == 0 {
		ui.PrintError("No services found in namespace %s", namespace)
		return nil, fmt.Errorf("no services found")
	}

	options := make([]string, len(services))
	for i, svc := range services {
		options[i] = ui.FormatServiceOption(&svc)
	}

	selected, err := ui.SelectFromOptions("🔍 Select service:", options)
	if err != nil {
		return nil, err
	}

	return &services[selected], nil
}

func selectPortInteractive(svc *corev1.Service) (int32, error) {
	if len(svc.Spec.Ports) == 0 {
		ui.PrintError("Service %s has no ports defined", svc.Name)
		return 0, fmt.Errorf("no ports defined")
	}

	if len(svc.Spec.Ports) == 1 {
		port := svc.Spec.Ports[0]
		fmt.Printf("\n⚡ Using port ")
		ui.NewPrinter().PrintPort(1, ui.FormatPortOption(port))
		return port.Port, nil
	}

	options := make([]string, len(svc.Spec.Ports))
	for i, port := range svc.Spec.Ports {
		options[i] = ui.FormatPortOption(port)
	}

	selected, err := ui.SelectFromOptions("🔌 Select port:", options)
	if err != nil {
		return 0, err
	}

	return svc.Spec.Ports[selected].Port, nil
}

func runServicePortForward(cmd *cobra.Command, args []string) error {
	client, err := k8s.NewClient()
	if err != nil {
		ui.PrintError("Error creating kubernetes client: %v", err)
		return err
	}

	var service *corev1.Service

	if len(args) == 0 {
		service, err = selectServiceInteractive(client, svcNamespace)
	} else {
		service, err = client.GetService(svcNamespace, args[0])
	}

	if err != nil {
		return err
	}

	if !cmd.Flags().Changed("port") {
		port, err := selectPortInteractive(service)
		if err != nil {
			return err
		}
		svcRemotePort = int(port)
	}

	if !cmd.Flags().Changed("local-port") {
		svcLocalPort, err = k8s.GetAvailablePort()
		if err != nil {
			ui.PrintError("Error getting random port: %v", err)
			return err
		}
	}

	opts := k8s.PortForwardOptions{
		Namespace:   svcNamespace,
		LocalPort:   svcLocalPort,
		RemotePort:  svcRemotePort,
		ServiceName: service.Name,
	}

	return client.ForwardPort(opts)
}
