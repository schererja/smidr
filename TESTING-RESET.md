# Testing Agent Reset-Enrollment

This guide shows how to test the new `reset-enrollment` command.

## Build the Agent

```bash
cd /Users/schererja/src/github.com/schererja/smidr/agent
go build -o ../bin/smidr-agent ./cmd/agent
```

## Run Tests

```bash
# Run the reset enrollment tests
go test -v ./internal/agent -run TestReset

# Run all agent tests
go test -v ./internal/agent
```

## Manual Testing

### 1. Check Current State

```bash
# View your current config
cat ~/Library/Application\ Support/smidr/agent.yaml

# Check if cert exists
ls -la ~/Library/Application\ Support/smidr/*.pem
```

### 2. Stop the Agent

```bash
# If running as daemon, stop it first
# (or just kill the process if running in foreground)
```

### 3. Run Reset Command

```bash
# Using default config path (macOS: ~/Library/Application Support/smidr/agent.yaml)
../bin/smidr-agent reset-enrollment

# Or specify config path
../bin/smidr-agent reset-enrollment --config ~/Library/Application\ Support/smidr/agent.yaml
```

Expected output:
```
Enrollment state reset successfully
Deleted certificate files and cleared agent ID from config
Run 'smidr-agent daemon' to re-enroll
```

### 4. Verify Reset

```bash
# Check config - agent_id and paths should be empty
cat ~/Library/Application\ Support/smidr/agent.yaml

# Check that cert/key/csr files are deleted
ls -la ~/Library/Application\ Support/smidr/*.pem
# Should show: No such file or directory
```

### 5. Re-enroll

```bash
# Start daemon - it will auto-generate new credentials and re-enroll
../bin/smidr-agent daemon
```

Expected behavior:
1. Agent generates new UUID for agent_id
2. Agent creates new RSA key pair
3. Agent generates new CSR
4. Agent POSTs to control plane /api/agents/register
5. Agent receives and saves new certificate
6. Agent starts heartbeat loop with new identity

### 6. Verify Re-enrollment

```bash
# Check config - should have new agent_id
cat ~/Library/Application\ Support/smidr/agent.yaml

# Check that new cert exists
ls -la ~/Library/Application\ Support/smidr/agent.crt.pem

# Check cert details (should have new agent ID in CN)
openssl x509 -in ~/Library/Application\ Support/smidr/agent.crt.pem -noout -subject -dates
```

## Use Case: Control Plane Database Reset

This is the exact scenario you encountered. The agent now automatically detects this:

```bash
# Scenario:
# 1. Agent enrolled with old database (ID: faf726fd-ccda-4cf6-971f-eafd039d65f3)
# 2. Control plane database was deleted and recreated
# 3. Agent tries to heartbeat but gets 404 "Agent not registered"

# What happens now:
# When agent daemon starts, it checks certificate validity on first heartbeat
# If 404 is returned, the daemon exits with clear error message:

# Agent log output:
# ERROR agent certificate is invalid - this usually happens when the control plane database was reset
# ERROR to fix this issue, run: smidr-agent reset-enrollment
# ERROR then restart the daemon to re-enroll with a new certificate

# Solution:
../bin/smidr-agent reset-enrollment
../bin/smidr-agent daemon

# Agent now has fresh identity and can successfully heartbeat
```

## Files Changed

- `agent/cmd/agent/main.go` - Added reset-enrollment command
- `agent/internal/agent/reset.go` - ResetEnrollment function
- `agent/internal/agent/reset_test.go` - Unit tests
- `agent/README.md` - Documentation updates

## What Gets Deleted

- `agent.crt.pem` - Certificate file
- `agent.csr.pem` - CSR file  
- `agent.key.pem` - Private key file
- Config fields: agent_id, key_path, csr_path, cert_path

## What Gets Preserved

- `control_plane_url` - Control plane endpoint
- `hostname` - System hostname
- `token` - Enrollment token (if set)
- `ca_cert_path` - CA certificate path (if set)
- `heartbeat_interval` - Heartbeat interval setting
