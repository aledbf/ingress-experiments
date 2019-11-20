package main

import (
	"context"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/apiserver/pkg/storage/storagebackend/factory"
)

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
	etcdStorage, df, err := factory.Create(storagebackend.Config{
		Type: storagebackend.StorageTypeETCD3,
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

		if wf != nil {
			log.Printf("Event: %v\n", wf.Data)
		}
	}

	log.Printf("done\n")
}
