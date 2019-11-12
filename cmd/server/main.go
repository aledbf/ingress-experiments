package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"k8s.io/klog"

	"github.com/aledbf/ingress-experiments/internal/pkg/signal"
	"github.com/r3labs/sse"
)

const (
	eventChannel = "messages"
)

func main() {

	klog.InitFlags(nil)
	defer klog.Flush()

	flag.Set("alsologtostderr", "true")
	flag.Parse()

	ctx := signal.SetupSignalHandler(context.Background())

	klog.Info("Starting SSE server...")
	srv := newSSEServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/events", srv.clientInformation(srv.HTTPHandler))

	// TODO: remove example
	runExamples(srv)

	httpServer := newHTTPServer(8080, mux)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				return
			}

			klog.Fatalf("Starting HTTP Server: %v", err)
		}
	}()

	<-ctx.Done()

	klog.Info("Shutting down SSE server...")
	srv.Publish(eventChannel, &sse.Event{
		Event: []byte("maintenance"),
		Data:  []byte(`{"reason": "server shutdown"}`),
	})

	time.Sleep(time.Second)
	srv.Close()

	klog.Info("Shutting down HTTP server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	httpServer.Shutdown(ctx)

	klog.Info("done")
	os.Exit(0)
}

func newHTTPServer(port int, mux *http.ServeMux) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%v", port),
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      300 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}
