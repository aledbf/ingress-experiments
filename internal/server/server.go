package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"syscall"
	"time"

	"k8s.io/apiserver/pkg/server/healthz"
	"k8s.io/klog"
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
  - restart connection
  - act on changes
  - send status periodically

*/

type Instance struct {
	cfg *Configuration

	ctx context.Context

	server *http.Server
}

func New(cfg *Configuration) *Instance {
	r := &Instance{
		cfg: cfg,
	}

	return r
}

func (r *Instance) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	registerHealthz(mux)
	registerHandlers(mux)

	server := newHTTPServer(r.cfg.ListenPort, mux)
	go func() {
		klog.Fatal(server.ListenAndServeTLS(r.cfg.Certificate, r.cfg.Key))
	}()
	r.server = server

	go func() {
		for {
			time.Sleep(10 * time.Second)
			fmt.Println("sleeping...")
		}
	}()

	return nil
}

func (r *Instance) Stop() {
	err := r.server.Shutdown(context.Background())
	if err != nil {
		klog.Warningf("Unexpected error stopping HTTP server: %v", err)
	}
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

func registerHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/build", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		//w.Write([]byte(version.String()))
		w.Write([]byte("0.0.0"))
	})

	mux.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`sending TERM signal`))
		err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		if err != nil {
			klog.Errorf("Unexpected error: %v", err)
		}
	})
}

func registerHealthz(mux *http.ServeMux) {
	healthz.InstallHandler(mux,
		healthz.PingHealthz,
	)
}
