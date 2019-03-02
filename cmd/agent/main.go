package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/aledbf/ingress-experiments/internal/agent"
	"github.com/aledbf/ingress-experiments/internal/signal"
)

const (
	ingressControllerURL = "ingress-controller-url"
	certificate          = "certificate"
	key                  = "key"
	podIP                = "pod-ip"
	podName              = "pod-name"
)

func main() {
	var cfg agent.Configuration

	runCommand := func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting NGINX ingress controller agent")
		a := agent.NewInstance(&cfg)

		contextCtx := signal.SigTermCancelContext(context.Background())
		if err := a.Run(contextCtx); err != nil {
			fmt.Printf("Unexpected error starting the agent: %v", err)
			os.Exit(1)
		}

		<-contextCtx.Done()
	}

	rootCmd := &cobra.Command{
		Use:   "agent",
		Short: "Run NGINX ingress controller agent",
		Run:   runCommand,
	}

	rootCmd.Flags().StringVarP(&cfg.ServerURL, ingressControllerURL, "", "", "URL of the NGINX ingress controller")
	rootCmd.MarkFlagRequired(ingressControllerURL)

	rootCmd.Flags().StringVarP(&cfg.Certificate, certificate, "", "", "TLS certificate fot mTLS")
	rootCmd.MarkFlagRequired(certificate)

	rootCmd.Flags().StringVarP(&cfg.Key, key, "", "", "TLS key fot mTLS")
	rootCmd.MarkFlagRequired(key)

	rootCmd.Flags().StringVarP(&cfg.PodIP, podIP, "", "", "IP address of the Pod where the agent is running")
	rootCmd.Flags().StringVarP(&cfg.PodName, podName, "", "", "Name of the Pod running the agent")

	rootCmd.PersistentFlags().BoolVarP(&cfg.Debug, "debug", "d", false, "Enable debug mode")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
