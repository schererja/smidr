# Binaries and paths
BIN_DIR := bin
AGENT_BIN := $(BIN_DIR)/smidr-agent
SERVER_BIN := $(BIN_DIR)/smidr-server
PKG_AGENT := ./cmd/smidr-agent
PKG_SERVER := ./cmd/smidr-core      # or your server package
GOFLAGS :=
LDFLAGS :=

.PHONY: all build test lint clean agent server run-agent run-server deploy-agent

all: build

## Build
build: agent server

agent:
	@mkdir -p $(BIN_DIR)
	GOFLAGS="$(GOFLAGS)" go build -o $(AGENT_BIN) -ldflags="$(LDFLAGS)" $(PKG_AGENT)

server:
	@mkdir -p $(BIN_DIR)
	GOFLAGS="$(GOFLAGS)" go build -o $(SERVER_BIN) -ldflags="$(LDFLAGS)" $(PKG_SERVER)

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
deploy-agent: agent
	ssh -t ik8ladmin@smidr-server.ik8labs.local 'sudo mkdir -p /opt/smidr/bin && sudo chown ik8ladmin:ik8ladmin /opt/smidr/bin'
	scp $(AGENT_BIN) ik8ladmin@smidr-server.ik8labs.local:/opt/smidr/bin/smidr-agent
	scp systemd/smidr-agent.service ik8ladmin@smidr-server.ik8labs.local:/tmp/smidr-agent.service
	ssh -t ik8ladmin@smidr-server.ik8labs.local 'sudo mv /tmp/smidr-agent.service /etc/systemd/system/smidr-agent.service && sudo systemctl daemon-reload && sudo systemctl enable --now smidr-agent'

## Clean
clean:
	rm -rf $(BIN_DIR)
