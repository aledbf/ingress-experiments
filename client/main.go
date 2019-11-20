package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coreos/etcd/clientv3"
	cli "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	grpcprom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/signalfx/embetcd/embetcd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

func main() {
	clientv3.SetLogger(grpclog.NewLoggerV2(os.Stderr, os.Stderr, os.Stderr))

	dialOptions := []grpc.DialOption{
		grpc.WithBlock(), // block until the underlying connection is up
		grpc.WithBackoffMaxDelay(100 * time.Millisecond),
		grpc.WithUnaryInterceptor(grpcprom.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpcprom.StreamClientInterceptor),
	}

	ctx := context.Background()
	client, err := embetcd.NewClient(cli.Config{
		Context:   ctx,
		Endpoints: []string{"http://127.0.0.1:2470"},

		DialOptions: dialOptions,

		MaxCallSendMsgSize: 10000000,

		//Username:    "root",
		//Password:    "123",
	})
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	go func() {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			_, err = client.Put(ctx, "/demo", time.Now().String())
			if err != nil {
				switch err {
				case context.Canceled:
					log.Fatalf("ctx is canceled by another routine: %v", err)
				case rpctypes.ErrEmptyKey:
					log.Fatalf("client-side error: %v", err)
				}
			}

			cancel()
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			_, err = client.Delete(ctx, "/demo")
			if err != nil {
				switch err {
				case context.Canceled:
					log.Fatalf("ctx is canceled by another routine: %v", err)
				case rpctypes.ErrEmptyKey:
					log.Fatalf("client-side error: %v", err)
				}
			}

			cancel()
			time.Sleep(11 * time.Second)
		}
	}()

	go func() {
		for {
			log.Printf("Start watch")
			watchChan := client.Watch(context.Background(), "/demo")
			for watchResp := range watchChan {
				for _, watchEvent := range watchResp.Events {
					k := string(watchEvent.Kv.Key)
					v := string(watchEvent.Kv.Value)
					version := watchEvent.Kv.Version

					log.Printf("%v: %v - %v (%v)", watchEvent.Type, k, v, version)
				}
			}

			log.Printf("End watch")
		}
	}()

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-term:
			log.Println("Received SIGTERM, cancelling")
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	log.Println("Waiting...")
	<-term
}
