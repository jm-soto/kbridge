package k8s

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

func NewClient() (*Client, error) {
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating clientset: %v", err)
	}

	return &Client{
		clientset: clientset,
		config:    config,
	}, nil
}

func (c *Client) ListPods(namespace string, labels []string) ([]corev1.Pod, error) {
	labelSelector := buildLabelSelector(labels)

	pods, err := c.clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("error listing pods: %v", err)
	}

	return pods.Items, nil
}

func (c *Client) ListServices(namespace string, labels []string) ([]corev1.Service, error) {
	labelSelector := buildLabelSelector(labels)

	services, err := c.clientset.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("error listing services: %v", err)
	}

	return services.Items, nil
}

func (c *Client) GetService(namespace, name string) (*corev1.Service, error) {
	service, err := c.clientset.CoreV1().Services(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting service: %v", err)
	}
	return service, nil
}

func buildLabelSelector(labels []string) string {
	return strings.Join(labels, ",")
}
