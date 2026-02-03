.PHONY: proto-gen proto-clean build build-control-plane build-agent run-control-plane run-agent clean

proto-gen:
	@echo "Generating proto files..."
	protoc \
		--proto_path=api/proto \
		--go_out=pkg \
		--go_opt=paths=source_relative \
		--go-grpc_out=pkg \
		--go-grpc_opt=paths=source_relative \
		agent/v1/agent.proto

build-control-plane: proto-gen
	@mkdir -p bin
	go build -o bin/control-plane ./cmd/control-plane

build-agent: proto-gen
	@mkdir -p bin
	go build -o bin/agent ./cmd/agent

build: build-control-plane build-agent

run-control-plane: proto-gen
	go run ./cmd/control-plane

run-agent: proto-gen
	go run ./cmd/agent

clean:
	rm -rf pkg/proto bin

help:
	@echo "Available targets:"
	@echo "  proto-gen            - Generate proto files"
	@echo "  build                - Build everything"
	@echo "  build-control-plane  - Build control plane"
	@echo "  build-agent          - Build agent"
	@echo "  run-control-plane    - Run control plane"
	@echo "  run-agent            - Run agent"
	@echo "  clean                - Clean build artifacts"
