
all: agent server

agent:fmt vet
	go build -o bin/agent github.com/aledbf/ingress-experiments/cmd/agent

server:fmt vet
	go build -o bin/server github.com/aledbf/ingress-experiments/cmd/server

run-agent: fmt vet
	go run ./cmd/agent/main.go

run-server: fmt vet
	go run ./cmd/server/main.go

fmt:
	go fmt ./internal/... ./cmd/...

vet:
	go vet ./internal/... ./cmd/...
