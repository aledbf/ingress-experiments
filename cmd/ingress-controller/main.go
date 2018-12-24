package main

import (
	"flag"
	"io"
	"net"
	"time"

	pb "github.com/aledbf/ingress-experiments/api/grpc"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"k8s.io/klog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	name               string
	notificationStream map[string]pb.Notification_SubscribeServer
}

func newServer() *server {
	return &server{
		name:               "notification",
		notificationStream: make(map[string]pb.Notification_SubscribeServer),
	}
}

func (s *server) Subscribe(stream pb.Notification_SubscribeServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		klog.Infof("Received a Subscription Request (%s, %s)", in.GetPodName(), in.GetPodUuid())
		_, ok := s.notificationStream[in.GetPodName()+":"+in.GetPodUuid()]
		if ok {
			klog.Infof("Client %s already subscribed", in.GetPodName())
			continue
		}

		s.notificationStream[in.GetPodName()+":"+in.GetPodUuid()] = stream
	}

	return nil
}

func (s *server) pushUpdates() {
	time.Sleep(1 * time.Second)

	for {
		for k, v := range s.notificationStream {
			klog.Infof("Sending update to client %v\n", k)
			if err := v.Send(&pb.Update{
				Latest: &timestamp.Timestamp{Nanos: int32(time.Now().Nanosecond())},
			}); err != nil {
				klog.Infof("Send failed %v\n", err)
			}

			time.Sleep(5 * time.Second)
		}
	}
}

func main() {
	klog.InitFlags(nil)
	defer klog.Flush()

	flag.Set("alsologtostderr", "true")
	flag.Parse()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		klog.Fatalf("failed to listen: %v", err)
	}

	klog.Info("Starting grpc server...")
	grpcServer := grpc.NewServer()
	myServer := newServer()
	pb.RegisterNotificationServer(grpcServer, myServer)

	go myServer.pushUpdates()

	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		klog.Fatalf("failed to serve: %v", err)
	}
}
