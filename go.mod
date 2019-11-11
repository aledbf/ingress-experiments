module github.com/aledbf/ingress-experiments

go 1.13

require (
	github.com/coreos/etcd v3.3.15+incompatible
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.12.1 // indirect
	github.com/signalfx/embetcd v0.0.9
	github.com/smartystreets/goconvey v1.6.4 // indirect
	go.etcd.io/etcd v3.3.12+incompatible
	google.golang.org/grpc v1.24.0
	k8s.io/apimachinery v0.0.0-20191109100838-fee41ff082ed
	k8s.io/apiserver v0.0.0-20191109104011-687a3dde5a6b
	k8s.io/client-go v0.0.0-20191109102209-3c0d1af94be5
	k8s.io/kubectl v0.0.0-20191109120237-d44ed977fb78
	sigs.k8s.io/controller-runtime v0.3.0
)

replace go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738
