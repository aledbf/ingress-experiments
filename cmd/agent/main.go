package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aledbf/ingress-experiments/internal/agent"
	"github.com/aledbf/ingress-experiments/internal/common"
	"k8s.io/klog"

	"github.com/spf13/cobra"
)

func main() {
	klog.InitFlags(nil)

	const (
		ingressControllerURL = "ingress-controller-url"
		certificate          = "certificate"
		key                  = "key"
		podIP                = "pod-ip"
		podName              = "pod-name"
	)

	var cfg common.AgentConfiguration

	runCommand := func(cmd *cobra.Command, args []string) {
		a, err := agent.NewRunCommand(&cfg)
		if err != nil {
			klog.Errorf("Unexpected error starting the agent: %v", err)
			os.Exit(1)
		}

		klog.Info("Starting NGINX ingress controller")
		contextCtx := sigTermCancelContext(context.Background())
		if err := a.Run(contextCtx); err != nil {
			klog.Errorf("Unexpected error starting the agent: %v", err)
			os.Exit(1)
		}

		<-contextCtx.Done()

		time.Sleep(common.ShutdownTimeout)
		os.Exit(0)
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
		klog.Error(err)
		os.Exit(1)
	}
}

func sigTermCancelContext(ctx context.Context) context.Context {
	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		select {
		case <-term:
			klog.Infof("Received SIGTERM, cancelling")
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx
}
