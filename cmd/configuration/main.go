package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	api "k8s.io/kubernetes/pkg/apis/core"

	"github.com/aledbf/ingress-experiments/internal/admission"
	configurationv1alpha1 "github.com/aledbf/ingress-experiments/internal/apis/clientset/versioned"
	v1alpha1 "github.com/aledbf/ingress-experiments/internal/apis/configuration/v1alpha1"
)

func main() {
	var (
		flags = pflag.NewFlagSet("", pflag.ExitOnError)

		configuration = flags.String("configuration", "", `Name of the Configuration containing custom global configurations for the controller.`)
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

	config, err := getConfiguration(*configuration, configurationvClient)
	if err != nil {
		log.Fatalf("unexpected error getting configuration %v: %v", *configuration, err)
	}
	log.Printf("%v\n", config)

	watchConfiguration(config.Name, config.Namespace, configurationvClient)

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

func getConfiguration(config string, client *configurationv1alpha1.Clientset) (*v1alpha1.Configuration, error) {
	if config == "" {
		// using default configuration
		return nil, nil
	}

	ns, name, err := cache.SplitMetaNamespaceKey(config)
	if err != nil {
		return nil, fmt.Errorf("invalid ns: %v", err)
	}

	if ns == "" {
		return nil, fmt.Errorf("invalid ns")
	}

	if name == "" {
		return nil, fmt.Errorf("invalid name")
	}

	return client.Configuration().Configurations(ns).Get(name, metav1.GetOptions{})
}

func watchConfiguration(ns, name string, client *configurationv1alpha1.Clientset) {
	opts := metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(api.ObjectNameField, name).String(),
	}
	watcher, err := client.Configuration().Configurations(ns).Watch(opts)
	if err != nil {
		log.Printf("%v\n", err)
	}

	got := <-watcher.ResultChan()
	log.Printf("%v\n", got)
}
