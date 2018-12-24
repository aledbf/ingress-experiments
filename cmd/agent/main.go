//
// client.go
//
// gRPC Push Notification Client
//

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"time"

	pb "github.com/aledbf/ingress-experiments/api/grpc"

	"google.golang.org/grpc"
	"k8s.io/klog"
)

var name string

const (
	address = "localhost:50051"
)

func recvNotification(stream pb.Notification_SubscribeClient) {
	for {
		resp, err := stream.Recv()
		if err != nil {
			klog.Errorf("failed to recv %v", err)
			break
		}
		if err == io.EOF {
			break
		}
		klog.Info(resp, err)
	}
}

func subscribe(client pb.NotificationClient, closeCh chan struct{}) {
	stream, err := client.Subscribe(context.Background())
	if err != nil {
		klog.Fatalf("failed to subscribe %v", err)
	}

	go recvNotification(stream)

	if err := stream.Send(&pb.Client{PodName: name, PodUuid: "0000000"}); err != nil {
		klog.Fatalf("unexpected error: %v", err)
	}

	<-closeCh
	stream.CloseSend()
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	klog.InitFlags(nil)
	defer klog.Flush()

	flag.Set("alsologtostderr", "true")
	flag.Parse()

	klog.Info("Starting grpc client...")

	name = fmt.Sprintf("%s:%d", "Client", rand.Intn(50))

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		klog.Fatalf("did not connect: %v", err)
	}

	client := pb.NewNotificationClient(conn)

	closeCh := make(chan struct{})

	subscribe(client, closeCh)
}
