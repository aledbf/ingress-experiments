package agent

import (
	"fmt"
	"time"

	"github.com/r3labs/sse"
)

// ConnectionCallbacks are callbacks that are triggered during the lifecycle
// of the connection to the ingress controller.
// These are invoked asynchronously.
type ConnectionCallbacks struct {
	// OnDisconnect is called when the connection with the ingress controller
	// is lost
	OnDisconnect func()
	// OnReconnect is called when the connection with the ingress controller
	// is revered
	OnReconnect func(float64)
	// OnData is called when there is a new SSE event
	OnData func(*sse.Event)
}

// NewClient createa a new SSE connection to the ingress controller listening
func NewClient(podName, podUUID, channel, url string, closeCh chan struct{}, callbacks ConnectionCallbacks) *sse.Client {
	events := make(chan *sse.Event)
	disconnection := make(chan *time.Time)
	client := sse.NewClient(fmt.Sprintf("%v?pod_name=%v&pod_uuid=%v", url, podName, podUUID))
	client.SubscribeChan(channel, events)

	onDisconnect := func(c *sse.Client) {
		t := time.Now()
		disconnection <- &t
		callbacks.OnDisconnect()
	}
	client.OnDisconnect(onDisconnect)

	go func(client *sse.Client, events chan *sse.Event, disconnection chan *time.Time) {
		var disconnectedSince *time.Time

		for {
			select {
			case event := <-events:
				if event.Data != nil && disconnectedSince != nil {
					callbacks.OnReconnect(time.Now().Sub(*disconnectedSince).Seconds())
					disconnectedSince = nil
				}
				callbacks.OnData(event)
			case t := <-disconnection:
				disconnectedSince = t
			case <-closeCh:
				client.Unsubscribe(events)
				break
			}
		}
	}(client, events, disconnection)

	return client
}
