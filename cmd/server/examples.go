package main

import (
	"time"

	"github.com/r3labs/sse"
)

func runExamples(srv *server) {
	// every two seconds send information about the latest
	// configuration available to connected agents.
	// TODO: remove example
	go func(srv *server) {
		for {
			srv.Publish(eventChannel, &sse.Event{
				Event: []byte("configuration"),
				Data:  []byte("<id of latest configuration>"),
			})
			time.Sleep(2 * time.Second)
		}
	}(srv)

	// TODO: remove example
	go func(srv *server) {
		for {
			clients := []string{}
			for k := range srv.connectedClients {
				clients = append(clients, k)
			}

			data, _ := serialize(clients)
			srv.Publish(eventChannel, &sse.Event{
				Event: []byte("agents"),
				Data:  data,
			})
			time.Sleep(5 * time.Second)
		}
	}(srv)
}
