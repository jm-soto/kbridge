package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kbridge",
	Short: "A tool to create port-forwards to Kubernetes resources",
	Long: `kbridge is a tool that allows you to create port-forwards 
to different Kubernetes resources like pods and services by specifying 
labels to filter the resources.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(newPodCmd())
	rootCmd.AddCommand(newServiceCmd())
}
