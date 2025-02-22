package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jm-soto/kbridge/pkg/k8s"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

var (
	podNamespace  string
	podLabels     []string
	podLocalPort  int
	podRemotePort int
)

func newPodCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pod",
		Short: "Port forward to a pod",
		Long: `Port forward to a pod selected by labels. If multiple pods match 
the labels, you will be prompted to select one.`,
		RunE: runPodPortForward,
	}

	cmd.Flags().StringVarP(&podNamespace, "namespace", "n", "default", "The namespace where the pod is located")
	cmd.Flags().StringArrayVarP(&podLabels, "label", "l", []string{}, "Labels to filter pods (can be specified multiple times)")
	cmd.Flags().IntVarP(&podLocalPort, "local-port", "L", 0, "Local port to forward from (random if not specified)")
	cmd.Flags().IntVarP(&podRemotePort, "port", "p", 80, "Remote port to forward to")

	cmd.MarkFlagRequired("label")

	return cmd
}

func selectPod(pods []corev1.Pod) (*corev1.Pod, error) {
	if len(pods) == 1 {
		return &pods[0], nil
	}

	fmt.Println("\nMultiple pods found. Please select one:")
	for i, pod := range pods {
		fmt.Printf("[%d] %s (Status: %s)\n", i+1, pod.Name, pod.Status.Phase)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nEnter the number of the pod (or 'q' to quit): ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if text == "q" {
			return nil, fmt.Errorf("user cancelled selection")
		}

		num, err := strconv.Atoi(text)
		if err != nil || num < 1 || num > len(pods) {
			fmt.Println("Invalid selection. Please try again.")
			continue
		}

		return &pods[num-1], nil
	}
}

func runPodPortForward(cmd *cobra.Command, args []string) error {
	client, err := k8s.NewClient()
	if err != nil {
		return err
	}

	// Si el puerto local no fue especificado, obtener uno aleatorio
	if !cmd.Flags().Changed("local-port") {
		podLocalPort, err = k8s.GetAvailablePort()
		if err != nil {
			return fmt.Errorf("error getting random port: %v", err)
		}
	}

	// Listar pods
	pods, err := client.ListPods(podNamespace, podLabels)
	if err != nil {
		return err
	}

	if len(pods) == 0 {
		return fmt.Errorf("no pods found with labels %v in namespace %s", podLabels, podNamespace)
	}

	// Selección inicial del pod
	selectedPod, err := selectPod(pods)
	if err != nil {
		return err
	}

	fmt.Printf("\nSelected pod: %s\n", selectedPod.Name)

	// Configurar las opciones de port-forward
	opts := k8s.PortForwardOptions{
		Namespace:  podNamespace,
		Labels:     podLabels,
		LocalPort:  podLocalPort,
		RemotePort: podRemotePort,
	}

	// Iniciar el port-forward con reconexión automática
	return client.ForwardPort(opts)
}
