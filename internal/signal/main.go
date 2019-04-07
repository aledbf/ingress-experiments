package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/klog"
)

func SigTermCancelContext(ctx context.Context) context.Context {
	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		select {
		case <-term:
			klog.Infof("Received SIGTERM, cancelling")
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx
}
