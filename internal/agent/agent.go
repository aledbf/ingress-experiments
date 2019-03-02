package agent

import (
	"context"
	"time"
)

type Configuration struct {
	ServerURL string

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
	ctx context.Context

	ws interface{}
}

func NewInstance(cfg *Configuration) *Runtime {
	return &Runtime{}
}

func (r *Runtime) Run(ctx context.Context) error {

	go func() {
		<-ctx.Done()
		r.stop()
	}()

	go func() {
		time.Sleep(1 * time.Second)
		// send local state
	}()

	return nil
}

func (r *Runtime) stop() {}

func (r *Runtime) connect() {}

func (r *Runtime) onMessage() {
	// check message type
	// call action
}
