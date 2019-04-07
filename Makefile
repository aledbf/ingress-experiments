
all: agent server

agent:fmt vet
	go build -o bin/agent github.com/aledbf/ingress-experiments/cmd/agent

server: server
	go build -o bin/server github.com/aledbf/ingress-experiments/cmd/server

run-agent: agent
	bin/agent --certificate cert.pem --key key.pem --ingress-controller-url https://localhost:10254

run-server: server
	bin/server --certificate cert.pem --key key.pem

fmt:
	go fmt ./internal/... ./cmd/...

vet:
	go vet ./internal/... ./cmd/...

test-certificate:
	openssl req -newkey rsa:2048 \
		-new -nodes -x509 \
		-days 3650 \
		-out cert.pem \
		-keyout key.pem \
		-subj "/O=Acme Co/OU=Fake SSL Certificate/CN=localhost"