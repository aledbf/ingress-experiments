package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/etcd/embed"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/signalfx/embetcd/embetcd"
)

func main() {
	var (
		listenPeerPort    int
		advertisePeerPort int

		peers string
	)

	flag.IntVar(&listenPeerPort, "listen-peer-port", 2480, "Listen peer port")
	flag.IntVar(&advertisePeerPort, "advertise-peer-port", 2470, "Listen advertise peer port")
	flag.StringVar(&peers, "peers", "", "Peers")

	flag.Parse()

	server := embetcd.New()

	etcdCfg := embetcd.NewConfig()

	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir) // clean up
	etcdCfg.Dir = dir

	etcdCfg.ClusterName = "ingress-controller"
	etcdCfg.ClusterState = embed.ClusterStateFlagNew
	etcdCfg.PeerTLSInfo = transport.TLSInfo{
		//CertFile: "localhost.crt",
		//KeyFile:  "localhost.key",
		//CAFile:   "rootCA.crt",
		//ClientCertAuth: true,
	}

	etcdCfg.LPUrls = []url.URL{{Scheme: "http", Host: fmt.Sprintf("0.0.0.0:%v", listenPeerPort)}}
	etcdCfg.APUrls = []url.URL{{Scheme: "http", Host: fmt.Sprintf("127.0.0.1:%v", listenPeerPort)}}

	etcdCfg.ClientTLSInfo = transport.TLSInfo{
		//CertFile: "localhost.crt",
		//KeyFile:  "localhost.key",
		//CAFile:   "rootCA.crt",
		//ClientCertAuth: true,
	}

	etcdCfg.LCUrls = []url.URL{{Scheme: "http", Host: fmt.Sprintf("0.0.0.0:%v", advertisePeerPort)}}
	etcdCfg.ACUrls = []url.URL{{Scheme: "http", Host: fmt.Sprintf("127.0.0.1:%v", advertisePeerPort)}}

	if peers != "" {
		log.Printf("Configurando peers: %v\n", peers)
		ic := strings.Split(peers, ",")
		//ic = append(ic, fmt.Sprintf("http://127.0.0.1:%v", advertisePeerPort))

		etcdCfg.ClusterState = embed.ClusterStateFlagExisting
		etcdCfg.InitialCluster = ic
	}

	timeout, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	err = server.Start(timeout, etcdCfg)
	if err != nil {
		log.Fatal(err)
		cancel()
	}

	if server.IsRunning() {
		log.Println("Running...")
	}

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-term:
			log.Println("Received SIGTERM, cancelling")
			cancel()
		case <-ctx.Done():
		}
	}()

	log.Println("Waiting...")
	<-term
}
