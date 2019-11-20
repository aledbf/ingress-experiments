module github.com/aledbf/ingress-experiments

go 1.13

require (
	github.com/coreos/etcd v3.3.13+incompatible
	github.com/golang/groupcache v0.0.0-20180513044358-24b0969c4cb7 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/signalfx/embetcd v0.0.9
	github.com/smartystreets/goconvey v1.6.4 // indirect
	google.golang.org/grpc v1.23.1

	k8s.io/apimachinery v0.0.0-20191109100838-fee41ff082ed
	k8s.io/apiserver v0.0.0-20191109104011-687a3dde5a6b
)
