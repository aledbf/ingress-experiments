package main

import (
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	"github.com/aledbf/ingress-experiments/internal/gossip"
)

func main() {
	var (
		flags = pflag.NewFlagSet("", pflag.ExitOnError)

		key = flags.String("key", "", `encryption key`)

		server = flags.Bool("server", false, `run as server or listener`)

		members = flags.StringArray("members", []string{}, `comma seperated list of members`)

		bindAddress = flags.String("bind-address", "0.0.0.0", `bind address`)
		port        = flags.Int("port", 4567, `bind port`)
	)

	flag.Set("logtostderr", "true")

	flags.AddGoFlagSet(flag.CommandLine)
	flags.Parse(os.Args)

	// Workaround for this issue:
	// https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})

	ip := net.ParseIP(*bindAddress)

	agent, err := gossip.NewAgent(&gossip.Config{
		EncryptKey: *key,
		Server:     *server,

		BindAddress: ip,
		BindPort:    *port,

		StartJoin: *members,
	})
	if err != nil {
		glog.Fatal(err)
	}

	if *server {
		glog.Infof("starting server on port %v...", *port)

		agent.StartServer()

		go func() {
			for {
				time.Sleep(10 * time.Second)
				glog.Info("sending message")
				agent.Broadcast(&gossip.IngressConfiguration{
					NGINX: []byte("byte array with nginx.cfg"),
				})
			}
		}()
	} else {
		glog.Infof("starting client joining %v on port %v...", *members, *port)
		agent.Start()
	}

	handleSigterm()
}

func handleSigterm() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	<-signalChan
	glog.Infof("Received SIGTERM, shutting down")
	time.Sleep(5 * time.Second)
}
