package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/pkg/transport"
	"google.golang.org/grpc"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/etcd3"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/apiserver/pkg/storage/value"
)

// DestroyFunc is to destroy any resources used by the storage returned in Create() together.
type DestroyFunc func()

type runningCompactor struct {
	interval time.Duration
	cancel   context.CancelFunc
	client   *clientv3.Client
	refs     int
}

var (
	lock       sync.Mutex
	compactors = map[string]*runningCompactor{}
)

// startCompactorOnce start one compactor per transport. If the interval get smaller on repeated calls, the
// compactor is replaced. A destroy func is returned. If all destroy funcs with the same transport are called,
// the compactor is stopped.
func startCompactorOnce(c storagebackend.TransportConfig, interval time.Duration) (func(), error) {
	lock.Lock()
	defer lock.Unlock()

	key := fmt.Sprintf("%v", c) // gives: {[server1 server2] keyFile certFile caFile}
	if compactor, foundBefore := compactors[key]; !foundBefore || compactor.interval > interval {
		compactorClient, err := newETCD3Client(c)
		if err != nil {
			return nil, err
		}

		if foundBefore {
			// replace compactor
			compactor.cancel()
			compactor.client.Close()
		} else {
			// start new compactor
			compactor = &runningCompactor{}
			compactors[key] = compactor
		}

		ctx, cancel := context.WithCancel(context.Background())

		compactor.interval = interval
		compactor.cancel = cancel
		compactor.client = compactorClient

		etcd3.StartCompactor(ctx, compactorClient, interval)
	}

	compactors[key].refs++

	return func() {
		lock.Lock()
		defer lock.Unlock()

		compactor := compactors[key]
		compactor.refs--
		if compactor.refs == 0 {
			compactor.cancel()
			compactor.client.Close()
			delete(compactors, key)
		}
	}, nil
}

func newETCD3HealthCheck(c storagebackend.Config) (func() error, error) {
	// constructing the etcd v3 client blocks and times out if etcd is not available.
	// retry in a loop in the background until we successfully create the client, storing the client or error encountered

	clientValue := &atomic.Value{}

	clientErrMsg := &atomic.Value{}
	clientErrMsg.Store("etcd client connection not yet established")

	go wait.PollUntil(time.Second, func() (bool, error) {
		client, err := newETCD3Client(c.Transport)
		if err != nil {
			clientErrMsg.Store(err.Error())
			return false, nil
		}
		clientValue.Store(client)
		clientErrMsg.Store("")
		return true, nil
	}, wait.NeverStop)

	return func() error {
		if errMsg := clientErrMsg.Load().(string); len(errMsg) > 0 {
			return fmt.Errorf(errMsg)
		}
		client := clientValue.Load().(*clientv3.Client)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if _, err := client.Cluster.MemberList(ctx); err != nil {
			return fmt.Errorf("error listing etcd members: %v", err)
		}
		return nil
	}, nil
}

func newETCD3Client(c storagebackend.TransportConfig) (*clientv3.Client, error) {
	tlsInfo := transport.TLSInfo{
		CertFile:      c.CertFile,
		KeyFile:       c.KeyFile,
		TrustedCAFile: c.TrustedCAFile,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return nil, err
	}
	// NOTE: Client relies on nil tlsConfig
	// for non-secure connections, update the implicit variable
	if len(c.CertFile) == 0 && len(c.KeyFile) == 0 && len(c.TrustedCAFile) == 0 {
		tlsConfig = nil
	}

	dialOptions := []grpc.DialOption{
		grpc.WithBackoffMaxDelay(2 * time.Second),

		grpc.WithUnaryInterceptor(grpcprom.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpcprom.StreamClientInterceptor),
	}

	cfg := clientv3.Config{
		DialTimeout:          1 * time.Second,
		DialKeepAliveTime:    1 * time.Second,
		DialKeepAliveTimeout: 1 * time.Second,
		DialOptions:          dialOptions,
		Endpoints:            c.ServerList,
		TLS:                  tlsConfig,

		PermitWithoutStream: true,
	}

	return clientv3.New(cfg)
}

func newETCD3Storage(c storagebackend.Config) (storage.Interface, DestroyFunc, error) {
	stopCompactor, err := startCompactorOnce(c.Transport, c.CompactionInterval)
	if err != nil {
		return nil, nil, err
	}

	client, err := newETCD3Client(c.Transport)
	if err != nil {
		stopCompactor()
		return nil, nil, err
	}

	var once sync.Once
	destroyFunc := func() {
		// we know that storage destroy funcs are called multiple times (due to reuse in subresources).
		// Hence, we only destroy once.
		// TODO: fix duplicated storage destroy calls higher level
		once.Do(func() {
			stopCompactor()
			client.Close()
		})
	}
	transformer := c.Transformer
	if transformer == nil {
		transformer = value.IdentityTransformer
	}
	return etcd3.New(client, c.Codec, c.Prefix, transformer, c.Paging), destroyFunc, nil
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TimestampData struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Data string `json:"data"`
}

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

func init() {
	utilruntime.Must(AddToScheme(scheme))
}

func main() {
	etcdStorage, df, err := newETCD3Storage(storagebackend.Config{
		Paging: true,
		Type:   "etcd3",
		Transport: storagebackend.TransportConfig{
			ServerList: []string{"http://127.0.0.1:2470"},
		},
		Codec: codecs.LegacyCodec(externalGV),
	})
	if err != nil {
		log.Fatal(err)
	}

	defer df()

	go func() {
		for {

			td := &TimestampData{}
			err := etcdStorage.Get(context.Background(), "demo", "0", td, false)
			if err != nil {
				td := &TimestampData{
					Data: time.Now().String(),
				}
				err = etcdStorage.Create(context.Background(), "demo", td, nil, 0)
				if err != nil {
					log.Printf("Error: %v", err)
				}
			}

			err = etcdStorage.GuaranteedUpdate(context.Background(), "demo", td, false, nil,
				storage.SimpleUpdate(func(obj runtime.Object) (runtime.Object, error) {
					curr := obj.(*TimestampData)
					curr.Data = time.Now().String()
					return curr, nil
				}))
			if err != nil {
				log.Printf("Error: %v", err)
			}

			time.Sleep(1 * time.Second)
		}
	}()

	watcher, err := etcdStorage.Watch(context.Background(), "demo", "0", storage.Everything)
	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Stop()

	for next := range watcher.ResultChan() {
		wf, ok := next.Object.(*TimestampData)
		if !ok {
			log.Printf("Event type: %T\n", next.Object)
		}

		log.Printf("Event: %v\n", wf.Data)
	}

	log.Printf("done\n")
}
