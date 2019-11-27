package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
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
func NewClient(ctx context.Context, podName, podUUID, channel, url string, callbacks ConnectionCallbacks) *sse.Client {
	client := sse.NewClient(fmt.Sprintf("%v?pod_name=%v&pod_uuid=%v", url, podName, podUUID))
	// use a constant backoff to avoid increasing delay in reconnection if the
	// server/s is/are down for more than two minutes.
	client.ReconnectStrategy = backoff.NewConstantBackOff(2 * time.Second)

	// subscribe a channel
	events := make(chan *sse.Event)
	client.SubscribeChan(channel, events)

	// create a channel to measure the time the agent is disconnected
	disconnection := make(chan *time.Time)
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
					callbacks.OnReconnect(time.Since(*disconnectedSince).Seconds())
					disconnectedSince = nil
				}

				if len(event.Data) > 1 {
					callbacks.OnData(event)
				}

			case t := <-disconnection:
				disconnectedSince = t
			case <-ctx.Done():
				client.Unsubscribe(events)
				return
			}
		}
	}(client, events, disconnection)

	return client
}
