package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/aledbf/ingress-experiments/internal/admission"
	configurationv1alpha1 "github.com/aledbf/ingress-experiments/internal/apis/clientset/versioned"
)

func main() {
	var (
		flags = pflag.NewFlagSet("", pflag.ExitOnError)
	)

	flag.Set("logtostderr", "true")

	flags.AddGoFlagSet(flag.CommandLine)
	flags.Parse(os.Args)

	// Workaround for this issue:
	// https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})

	cfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.Fatalf("error building kubeconfig: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("error building kubernetes clientset: %v", err)
	}

	configurationvClient, err := configurationv1alpha1.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("failed to create configuration clientset: %v", err)
	}

	wh := admission.NewValidatingAdmissionWebhook("", kubeClient, configurationvClient)
	if err := wh.Register(); err != nil {
		log.Fatalf("failed to register admission webhook: %v", err)
	}
	go wh.Run(nil)

	handleSigterm()
}

func handleSigterm() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	<-signalChan
	glog.Infof("Received SIGTERM, shutting down")
	time.Sleep(5 * time.Second)
}
