package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/r3labs/sse"
	"k8s.io/klog"
)

const (
	eventChannel = "messages"
)

func main() {

	klog.InitFlags(nil)
	defer klog.Flush()

	flag.Set("alsologtostderr", "true")
	flag.Parse()

	klog.Info("Starting SSE server...")
	srv := newServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/events", srv.clientInformation(srv.sseServer.HTTPHandler))

	// every two seconds send information about the latest
	// configuration available to connected agents.
	go func(sseServer *sse.Server) {
		for {
			sseServer.Publish(eventChannel, &sse.Event{
				Event: []byte("configuration"),
				Data:  []byte("<id of latest configuration>"),
			})
			time.Sleep(2 * time.Second)
		}
	}(srv.sseServer)

	go func(srv *server) {
		for {
			clients := []string{}
			for k := range srv.connectedClients {
				clients = append(clients, k)
			}

			data, _ := serialize(clients)
			srv.sseServer.Publish(eventChannel, &sse.Event{
				Event: []byte("agents"),
				Data:  data,
			})
			time.Sleep(5 * time.Second)
		}
	}(srv)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		klog.Fatal(err)
	}
}
