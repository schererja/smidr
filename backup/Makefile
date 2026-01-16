# Binaries and paths
BIN_DIR := bin
AGENT_BIN := $(BIN_DIR)/smidr-agent
SERVER_BIN := $(BIN_DIR)/smidr-server
PKG_AGENT := ./cmd/smidr-agent
PKG_SERVER := ./cmd/smidr-core      # or your server package
GOFLAGS :=
LDFLAGS :=

# Deployment configuration - can be overridden via `make deploy-agent AGENT_USER=user AGENT_HOST=host`
AGENT_USER ?= ik8ladmin
AGENT_HOST ?= smidr-server.ik8labs.local
SERVER_USER ?= ik8ladmin
SERVER_HOST ?= smidr-server-agent.ik8labs.local

.PHONY: all build test lint clean agent server run-agent run-server deploy-agent deploy-server agent-linux server-linux build-linux

all: build

## Build
build: agent server

agent:
	@mkdir -p $(BIN_DIR)
	GOFLAGS="$(GOFLAGS)" go build -o $(AGENT_BIN) -ldflags="$(LDFLAGS)" $(PKG_AGENT)

server:
	@mkdir -p $(BIN_DIR)
	GOFLAGS="$(GOFLAGS)" go build -o $(SERVER_BIN) -ldflags="$(LDFLAGS)" $(PKG_SERVER)

## Cross-compile for Linux (from macOS/other)
LINUX_AGENT_BIN := $(BIN_DIR)/smidr-agent-linux-amd64
LINUX_SERVER_BIN := $(BIN_DIR)/smidr-server-linux-amd64

agent-linux:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(LINUX_AGENT_BIN) -ldflags="$(LDFLAGS)" $(PKG_AGENT)

server-linux:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(LINUX_SERVER_BIN) -ldflags="$(LDFLAGS)" $(PKG_SERVER)

build-linux: agent-linux server-linux

## Tests
test:
	go test ./...

## Lint (if using golangci-lint)
lint:
	golangci-lint run ./...

## Run locally
run-agent: agent
	./$(AGENT_BIN) -config ./config-agent.yaml

run-server: server
	./$(SERVER_BIN) -config ./config.yaml

## Deploy agent (copy to remote and restart)
deploy-agent: agent-linux
	@echo "Deploying agent to $(AGENT_USER)@$(AGENT_HOST)..."
	ssh -t $(AGENT_USER)@$(AGENT_HOST) 'sudo mkdir -p /opt/smidr/bin /etc/smidr /var/lib/smidr/work && sudo chown $(AGENT_USER):$(AGENT_USER) /opt/smidr/bin && sudo chown $(AGENT_USER):$(AGENT_USER) /var/lib/smidr/work'
	scp $(LINUX_AGENT_BIN) $(AGENT_USER)@$(AGENT_HOST):/opt/smidr/bin/smidr-agent
	scp config/agent-config.example.yaml $(AGENT_USER)@$(AGENT_HOST):/tmp/agent-config.yaml
	scp systemd/smidr-agent.service $(AGENT_USER)@$(AGENT_HOST):/tmp/smidr-agent.service
	ssh -t $(AGENT_USER)@$(AGENT_HOST) 'sudo mv /tmp/agent-config.yaml /etc/smidr/config-agent.yaml && sudo mv /tmp/smidr-agent.service /etc/systemd/system/smidr-agent.service && sudo systemctl daemon-reload && sudo systemctl enable --now smidr-agent'

## Deploy server (copy to remote and restart)
deploy-server: server-linux
	@echo "Deploying server to $(SERVER_USER)@$(SERVER_HOST)..."
	ssh -t $(SERVER_USER)@$(SERVER_HOST) 'sudo mkdir -p /opt/smidr/bin /etc/smidr /var/lib/smidr/data && sudo chown $(SERVER_USER):$(SERVER_USER) /opt/smidr/bin && sudo chown $(SERVER_USER):$(SERVER_USER) /var/lib/smidr/data'
	scp $(LINUX_SERVER_BIN) $(SERVER_USER)@$(SERVER_HOST):/tmp/smidr-server
	scp config/server-config.example.yaml $(SERVER_USER)@$(SERVER_HOST):/tmp/server-config.yaml
	scp systemd/smidr-server.service $(SERVER_USER)@$(SERVER_HOST):/tmp/smidr-server.service
	ssh -t $(SERVER_USER)@$(SERVER_HOST) 'sudo mv /tmp/smidr-server /opt/smidr/bin/smidr-server && sudo chmod +x /opt/smidr/bin/smidr-server && sudo mv /tmp/server-config.yaml /etc/smidr/config-server.yaml && sudo mv /tmp/smidr-server.service /etc/systemd/system/smidr-server.service && sudo systemctl daemon-reload && sudo systemctl enable --now smidr-server'

## Clean
clean:
	rm -rf $(BIN_DIR)
