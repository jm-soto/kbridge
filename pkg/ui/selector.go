package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func SelectFromOptions(prompt string, options []string) (int, error) {
	fmt.Printf("\n%s\n", prompt)

	for i, opt := range options {
		NewPrinter().PrintOption(i+1, opt)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nEnter number (q to quit): ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if text == "q" {
			return -1, fmt.Errorf("user cancelled selection")
		}

		num, err := strconv.Atoi(text)
		if err != nil || num < 1 || num > len(options) {
			PrintError("Invalid selection. Please try again.")
			continue
		}

		return num - 1, nil
	}
}

func FormatServiceOption(svc *corev1.Service) string {
	ports := []string{}
	for _, port := range svc.Spec.Ports {
		ports = append(ports, fmt.Sprintf("%d/%s", port.Port, port.Protocol))
	}
	return fmt.Sprintf("%s → %s", svc.Name, strings.Join(ports, ", "))
}

func FormatPortOption(port corev1.ServicePort) string {
	portInfo := fmt.Sprintf("%d/%s", port.Port, port.Protocol)
	if port.Name != "" {
		portInfo += fmt.Sprintf(" (%s)", port.Name)
	}
	return portInfo
}

func (p *Printer) PrintOption(index int, option string) {
	p.cyan.Printf("[%d] ", index)
	p.green.Printf("%s\n", option)
}
