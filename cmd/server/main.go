package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/klog"

	"github.com/aledbf/ingress-experiments/internal/server"
	"github.com/aledbf/ingress-experiments/internal/signal"
)

const (
	listenPort  = "listen-port"
	certificate = "certificate"
	key         = "key"
	podIP       = "pod-ip"
	podName     = "pod-name"
)

func main() {
	klog.InitFlags(nil)

	var cfg server.Configuration

	runCommand := func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting NGINX ingress controller")
		a := server.New(&cfg)

		contextCtx := signal.SigTermCancelContext(context.Background())
		if err := a.Run(contextCtx); err != nil {
			fmt.Printf("Unexpected error starting the agent: %v", err)
			os.Exit(1)
		}

		<-contextCtx.Done()

		a.Stop()

		time.Sleep(30 * time.Second)
		os.Exit(0)
	}

	rootCmd := &cobra.Command{
		Use:   "run",
		Short: "Run NGINX ingress controller",
		Run:   runCommand,
	}

	rootCmd.Flags().IntVarP(&cfg.ListenPort, listenPort, "", 10254, "URL of the NGINX ingress controller")

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
