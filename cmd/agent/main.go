package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/aledbf/ingress-experiments/internal/pkg/agent"
	"github.com/r3labs/sse"
	"k8s.io/klog"
)

const (
	eventChannel = "messages"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	klog.InitFlags(nil)
	defer klog.Flush()

	flag.Set("alsologtostderr", "true")
	flag.Parse()

	klog.Info("Starting SSE client...")

	podName := fmt.Sprintf("agent-%v", time.Now().Nanosecond())
	podUUID := "00001"

	closeCh := make(chan struct{})

	callbacks := agent.ConnectionCallbacks{
		OnDisconnect: func() {
			klog.Infof("ondisconnect")
		},
		OnReconnect: func(secondsOffline float64) {
			klog.Infof("Disconnected for %v seconds", secondsOffline)
		},
		OnData: func(event *sse.Event) {
			klog.Infof("Event type: %s - Data: '%s'", event.Event, event.Data)
		},
	}

	agent.NewClient(podName, podUUID,
		eventChannel,
		"http://localhost:8080/events",
		closeCh,
		callbacks)

	<-closeCh
}
