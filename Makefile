
.PHONY: all
all: build

.PHONY: build
build:
	go build \
		-a -installsuffix cgo \
		-ldflags '-s -w' \
		github.com/aledbf/ingress-experiments/cmd/configuration

.PHONY: dep-ensure
dep-ensure:
	dep version || go get -u github.com/golang/dep/cmd/dep
	dep ensure -v
	dep prune -v
	find vendor -name '*_test.go' -delete
