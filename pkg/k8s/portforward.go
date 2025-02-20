package k8s

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"kbridge/pkg/ui"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

const (
	maxRetries    = 5
	retryInterval = 5 * time.Second
)

type PortForwardOptions struct {
	Namespace   string
	Labels      []string
	LocalPort   int
	RemotePort  int
	ServiceName string // Si está vacío, se usa pod
}

func GetAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func (c *Client) ForwardPort(opts PortForwardOptions) error {
	var lastErr error
	retries := 0

	for {
		var err error
		if opts.ServiceName != "" {
			err = c.forwardToService(opts)
		} else {
			err = c.forwardToPod(opts)
		}

		if err != nil {
			lastErr = err
			retries++
			if retries > maxRetries {
				return fmt.Errorf("max retries (%d) exceeded, last error: %v", maxRetries, lastErr)
			}
			fmt.Printf("\nConnection lost. Retrying in %v... (Attempt %d/%d)\n", retryInterval, retries, maxRetries)
			time.Sleep(retryInterval)
			continue
		}

		return nil
	}
}

func (c *Client) forwardToService(opts PortForwardOptions) error {
	service, err := c.GetService(opts.Namespace, opts.ServiceName)
	if err != nil {
		ui.PrintError("Error getting service: %v", err)
		return err
	}

	selector := service.Spec.Selector
	if len(selector) == 0 {
		ui.PrintError("Service %s has no selector", opts.ServiceName)
		return fmt.Errorf("service has no selector")
	}

	labelSelector := formatLabelSelector(selector)
	pods, err := c.clientset.CoreV1().Pods(opts.Namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})

	if err != nil {
		ui.PrintError("Error listing pods: %v", err)
		return err
	}

	if len(pods.Items) == 0 {
		ui.PrintError("No pods found for service %s", opts.ServiceName)
		return fmt.Errorf("no pods found")
	}

	var targetPod *corev1.Pod
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			targetPod = &pod
			break
		}
	}

	if targetPod == nil {
		ui.PrintError("No running pods found for service %s", opts.ServiceName)
		return fmt.Errorf("no running pods")
	}

	return c.forwardPortToPod(targetPod, opts.LocalPort, opts.RemotePort)
}

func (c *Client) forwardToPod(opts PortForwardOptions) error {
	pods, err := c.ListPods(opts.Namespace, opts.Labels)
	if err != nil {
		return fmt.Errorf("error listing pods: %v", err)
	}

	var runningPods []corev1.Pod
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodRunning {
			runningPods = append(runningPods, pod)
		}
	}

	if len(runningPods) == 0 {
		return fmt.Errorf("no running pods found with labels %v", opts.Labels)
	}

	targetPod := &runningPods[0]
	return c.forwardPortToPod(targetPod, opts.LocalPort, opts.RemotePort)
}

func (c *Client) forwardPortToPod(pod *corev1.Pod, localPort, remotePort int) error {
	roundTripper, upgrader, err := spdy.RoundTripperFor(c.config)
	if err != nil {
		ui.PrintError("Error creating round tripper: %v", err)
		return err
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", pod.Namespace, pod.Name)
	serverURL := url.URL{Scheme: "https", Path: path, Host: strings.TrimPrefix(c.config.Host, "https://")}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL)

	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)

	ports := []string{fmt.Sprintf("%d:%d", localPort, remotePort)}
	pf, err := portforward.New(dialer, ports, stopChan, readyChan, out, errOut)
	if err != nil {
		ui.PrintError("Error creating port-forward: %v", err)
		return err
	}

	errChan := make(chan error, 1)
	go func() {
		if err := pf.ForwardPorts(); err != nil {
			errChan <- err
		}
	}()

	printer := ui.NewPrinter()

	select {
	case <-readyChan:
		printer.PrintForward(localPort, remotePort)
		printer.PrintLocalURL(localPort)
		printer.PrintExit()

		select {
		case err := <-errChan:
			ui.PrintError("Port forward error: %v", err)
			return err
		case <-stopChan:
			return nil
		}

	case err := <-errChan:
		close(stopChan)
		ui.PrintError("Port forward failed: %v", err)
		return err
	}
}

func formatLabelSelector(selector map[string]string) string {
	var selectors []string
	for k, v := range selector {
		selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(selectors, ",")
}
