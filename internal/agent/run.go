package agent

import (
	"context"
	"net/url"
	"time"

	"k8s.io/klog"

	"github.com/aledbf/ingress-experiments/internal/common"
	"github.com/aledbf/ingress-experiments/internal/network"
	"github.com/aledbf/ingress-experiments/internal/nginx"
)

type Instance struct {
	cfg *common.AgentConfiguration

	ngx *nginx.NGINX
}

func New(cfg *common.AgentConfiguration) (*Instance, error) {
	_, err := url.ParseRequestURI(cfg.ServerURL)
	if err != nil {
		return nil, err
	}

	return &Instance{
		cfg: cfg,
	}, nil
}

func (cmd *Instance) checkForUpdates() {
	update, ok := network.RequestConfiguration(cmd.cfg)
	if !ok {
		return
	}

	if update == nil {
		klog.Errorf("Update invalido")
		return
	}

	klog.Infof("Update received: cfg -> (%v) - lua -> (%v) - ssl -> (%v)",
		update.Configuration != nil,
		update.LUA != nil,
		update.Certificates != nil)

	err := cmd.ngx.Update(update)
	if err != nil {
		klog.Errorf("Unexpected error updating configuration: %v", err)
	}
}

func (cmd *Instance) process(ctx context.Context) {
	tick := time.Tick(common.CheckInterval)

	for {
		select {
		case <-tick:
			klog.Info("check")
			cmd.checkForUpdates()
		case <-ctx.Done():
			klog.Info("done")
			return
		}
	}
}

/*
func (cmd *Instance) serveMetrics(mux *http.ServeMux) {
	registry := prometheus.NewRegistry()
	// Metrics about API connections
	registry.MustRegister(cmd.networkRequestStatusesCollector)
	// Metrics about jobs failures
	registry.MustRegister(cmd.failuresCollector)
	// Metrics about the program's build version.
	registry.MustRegister(common.AppVersion.NewMetricsCollector())
	// Go-specific metrics about the process (GC stats, goroutines, etc.).
	registry.MustRegister(prometheus.NewGoCollector())
	// Go-unrelated process metrics (memory usage, file descriptors, etc.).
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
}
*/

func (cmd *Instance) Run(ctx context.Context) error {
	// start nginx
	go cmd.ngx.Start()

	go cmd.process(ctx)
	// start http healthz
	// start metrics
	return nil
}
