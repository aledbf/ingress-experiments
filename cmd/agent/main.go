package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/r3labs/sse"
	"k8s.io/klog/v2"

	"github.com/aledbf/ingress-experiments/internal/pkg/agent"
	"github.com/aledbf/ingress-experiments/internal/pkg/signal"
)

const (
	eventChannel = "messages"
)

func main() {
	klog.InitFlags(nil)
	defer klog.Flush()

	flag.Set("alsologtostderr", "true")
	flag.Parse()

	klog.Info("Starting SSE client...")

	podName := fmt.Sprintf("agent-%v", time.Now().Nanosecond())
	podUUID := "00001"

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

	ctx := signal.SetupSignalHandler(context.Background())

	agent.NewClient(ctx,
		podName, podUUID,
		eventChannel,
		"http://localhost:8080/events",
		callbacks)

	<-ctx.Done()

	time.Sleep(10 * time.Second)
	os.Exit(0)
}
