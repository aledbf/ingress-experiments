package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/k8s.io/kubernetes/pkg/healthz"
	"k8s.io/klog"

	"github.com/aledbf/ingress-experiments/internal/version"
)

type Configuration struct {
	ListenPort int

	Certificate string
	Key         string

	PodIP   string
	PodName string

	Debug bool
}

/*

TODO:
  - health-check
  - restart connection
  - act on changes
  - send status periodically
  -

*/

type Runtime struct {
	cfg *Configuration

	ctx context.Context

	hub *Hub
}

func NewInstance(cfg *Configuration) *Runtime {
	return &Runtime{
		cfg: cfg,
		hub: newHub(),
	}
}

func (r *Runtime) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	registerHealthz(mux)
	registerHandlers(r.hub, mux)

	go startHTTPServer(r.cfg.ListenPort, mux)
	go r.hub.run()

	go func() {
		for {
			time.Sleep(10 * time.Second)
			fmt.Println("sleeping...")
		}
	}()

	return nil
}

func (r *Runtime) stop() {}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		klog.Errorf("Unexpected error upgrading request: %v", err)
		return
	}

	agent := &Agent{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	agent.hub.register <- agent

	go agent.writePump()
	go agent.readPump()
}

func startHTTPServer(port int, mux *http.ServeMux) {
	server := &http.Server{
		Addr:              fmt.Sprintf(":%v", port),
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      300 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	klog.Fatal(server.ListenAndServe())
}

func registerProfiler(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/heap", pprof.Index)
	mux.HandleFunc("/debug/pprof/mutex", pprof.Index)
	mux.HandleFunc("/debug/pprof/goroutine", pprof.Index)
	mux.HandleFunc("/debug/pprof/threadcreate", pprof.Index)
	mux.HandleFunc("/debug/pprof/block", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

func registerHandlers(hub *Hub, mux *http.ServeMux) {
	mux.HandleFunc("/build", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(version.String()))
	})

	mux.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`sending TERM signal`))
		err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		if err != nil {
			klog.Errorf("Unexpected error: %v", err)
		}
	})

	mux.HandleFunc("/ws/v1", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
}

func registerHealthz(mux *http.ServeMux) {
	healthz.InstallHandler(mux,
		healthz.PingHealthz,
	)
}
