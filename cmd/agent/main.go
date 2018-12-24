package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

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

	events := make(chan *sse.Event)
	disconnection := make(chan *time.Time)

	podName := fmt.Sprintf("agent-%v", time.Now().Nanosecond())
	client := sse.NewClient(fmt.Sprintf("http://localhost:8080/events?pod_name=%v&pod_uuid=000001", podName))
	client.SubscribeChan(eventChannel, events)

	onDisconnect := func(c *sse.Client) {
		t := time.Now()
		disconnection <- &t
		klog.Infof("Disconnected: %v\n", t)
	}
	client.OnDisconnect(onDisconnect)

	go func(events chan *sse.Event, disconnection chan *time.Time) {
		var disconnectedSince *time.Time

		for {
			select {
			case event := <-events:
				klog.Infof("Event type: %s - Data: '%s'", event.Event, event.Data)
				if event.Data != nil && disconnectedSince != nil {
					klog.Infof("Disconnected for %v seconds\n", time.Now().Sub(*disconnectedSince).Seconds())
					disconnectedSince = nil
				}
			case t := <-disconnection:
				disconnectedSince = t
			}
		}
	}(events, disconnection)

	time.Sleep(10 * time.Hour)
}
