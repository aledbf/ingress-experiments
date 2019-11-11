package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/aledbf/ingress-experiments/internal/pkg/agent"
	"github.com/r3labs/sse"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

// ClientInfo defines information related to the client
type ClientInfo struct {
	Name string
	UUID string

	Pod *corev1.Pod
}

type server struct {
	sseServer *sse.Server

	mutex            *sync.Mutex
	connectedClients map[string]*ClientInfo
}

func newServer() *server {
	sseServer := sse.New()
	sseServer.AutoReplay = false
	sseServer.CreateStream(eventChannel)

	return &server{
		sseServer:        sseServer,
		mutex:            &sync.Mutex{},
		connectedClients: make(map[string]*ClientInfo, 0),
	}
}

func (s *server) clientInformation(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		klog.Info("Extracting client information")

		podName := r.URL.Query().Get("pod_name")
		if podName == "" {
			http.Error(w, "Query variable pod_name missing", 403)
			return
		}

		podUUID := r.URL.Query().Get("pod_uuid")
		if podUUID == "" {
			http.Error(w, "Query variable pod_uuid missing", 403)
			return
		}

		// validate the pod should be allowed to connect to the event stream
		err := s.addClient(podName, podUUID)
		if err != nil {
			http.Error(w, "Invalid client information", 403)
			return
		}

		defer s.removeClient(podName, podUUID)
		h.ServeHTTP(w, r)
	})
}

func (s *server) removeClient(podName, podUUID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.connectedClients[podName]; !ok {
		return
	}

	klog.Infof("Removing client: %v", podName)
	delete(s.connectedClients, podName)

	data, err := serialize(agent.NewEvent(agent.RemoveAction, podName))
	if err != nil {
		return
	}

	s.sseServer.Publish(eventChannel, &sse.Event{
		Event: []byte("agent"),
		Data:  data,
	})
}

func (s *server) addClient(podName, podUUID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.connectedClients[podName]; ok {
		return fmt.Errorf("Client %v is already connected", podName)
	}

	// TODO: validate podName and podUUID are allowed to connect.

	klog.Infof("Adding new client: %v", podName)
	s.connectedClients[podName] = &ClientInfo{
		Name: podName,
		UUID: podUUID,
	}

	data, err := serialize(agent.NewEvent(agent.AddAction, podName))
	if err != nil {
		return err
	}

	s.sseServer.Publish(eventChannel, &sse.Event{
		Event: []byte("agent"),
		Data:  data,
	})

	return nil
}

func serialize(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}
