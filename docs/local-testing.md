# Local Testing with LocalStack

## Overview

LocalStack emulates all AWS services locally, allowing you to test the entire Smidr architecture (agent, control plane, logging, job queue) before deploying to AWS.

**Stack**:
- **LocalStack**: API Gateway, Lambda, DynamoDB, SQS, S3, CloudWatch, Logs
- **DynamoDB Admin UI**: Visual browser for DynamoDB tables (optional, port 8001)

---

## Quick Start

### 1. Start LocalStack

```bash
cd /path/to/smidr
docker-compose up -d
```

This:
- Starts LocalStack on `http://localhost:4566`
- Initializes DynamoDB tables (Agents, Jobs, Logs, LogBatches)
- Creates SQS queues (JobQueue, JobQueue.fifo)
- Creates S3 buckets (smidr-artifacts, smidr-logs)
- Starts DynamoDB Admin UI on `http://localhost:8001`

### 2. Check LocalStack Health

```bash
# Verify services are running
docker-compose ps

# Verify LocalStack endpoint
curl -s http://localhost:4566/_localstack/health | jq .
```

### 3. Stop LocalStack

```bash
docker-compose down
```

---

## Agent Testing

### Configure Agent for Local Testing

Create `smidr-local.yaml`:

```yaml
agent:
  id: test-agent-001
  name: "local-test-agent"
  version: "dev"

control_plane:
  url: "http://localhost:8000"  # Local control plane (or localhost:4566 for direct LocalStack)
  tenant_id: "test-tenant"
  hmac_secret: "ZGV2LXRlc3Qtc2VjcmV0"  # base64("dev-test-secret")

logging:
  local_dir: "/tmp/smidr/logs"
  ring_buffer_mb: 10
  batch_interval_sec: 5
  batch_size_entries: 100
  retry_max_hours: 24

execution:
  runtime: "docker"
  max_concurrent_jobs: 2
  timeout_sec: 300

plugins:
  enabled: false
```

### Run Agent Against LocalStack

```bash
# Build agent
cd /path/to/smidr/agent
go build -o smidr-agent

# Run agent with local config
./smidr-agent -config ../smidr-local.yaml
```

The agent will:
1. Register with control plane at `http://localhost:8000`
2. Poll for jobs every 5 seconds
3. Buffer logs locally
4. Push log batches to `http://localhost:8000/api/jobs/{jobId}/logs/batch`

---

## Control Plane Testing

### Deploy Control Plane to LocalStack

Option A: **Run TypeScript handlers locally** (Node.js)

```bash
cd /path/to/smidr/control-plane
npm install
npm run dev  # Runs on http://localhost:8000
```

Option B: **Run handlers in LocalStack Lambda** (simulates AWS)

```bash
# Build Lambda bundle
npm run build

# Deploy to LocalStack
aws --endpoint-url http://localhost:4566 \
  lambda create-function \
  --function-name smidr-handlers \
  --runtime nodejs20.x \
  --role arn:aws:iam::000000000000:role/lambda-role \
  --handler index.handler \
  --zip-file fileb://dist/lambda.zip

# Invoke handler locally
aws --endpoint-url http://localhost:4566 \
  lambda invoke \
  --function-name smidr-handlers \
  --payload '{"action":"claimJob","agentId":"test-agent-001"}' \
  response.json
```

---

## End-to-End Test Scenarios

### Scenario 1: Agent Comes Online, Claims Job, Executes

1. **Start LocalStack**:
   ```bash
   docker-compose up -d
   ```

2. **Insert test job in DynamoDB**:
   ```bash
   aws --endpoint-url http://localhost:4566 dynamodb put-item \
     --table-name Jobs \
     --item '{
       "job_id": {"S": "job-001"},
       "created_at": {"S": "2025-01-01T00:00:00Z"},
       "agent_id": {"S": "test-agent-001"},
       "status": {"S": "Pending"},
       "job_type": {"S": "build"},
       "parameters": {"M": {"command": {"S": "echo hello"}}},
       "deadline_utc": {"S": "2025-01-02T00:00:00Z"}
     }'
   ```

3. **Agent claims job**:
   ```bash
   curl -X POST http://localhost:8000/api/agents/test-agent-001/jobs/claim \
     -H "X-Tenant-Id: test-tenant" \
     -H "X-Agent-Id: test-agent-001" \
     -H "X-Timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)" \
     -H "X-Signature: <HMAC-SHA256>" \
     -H "Content-Type: application/json"
   ```

4. **Agent executes job, pushes logs**:
   ```bash
   curl -X POST http://localhost:8000/api/jobs/job-001/logs/batch \
     -H "X-Tenant-Id: test-tenant" \
     -H "X-Agent-Id: test-agent-001" \
     -H "X-Timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)" \
     -H "X-Signature: <HMAC-SHA256>" \
     -H "Content-Type: application/json" \
     -d '{
       "job_id": "job-001",
       "build_id": "build-001",
       "logs": [
         {
           "timestamp": "2025-01-01T00:00:01Z",
           "level": "INFO",
           "sequence": 1,
           "source": "executor",
           "message": "Starting job execution"
         },
         {
           "timestamp": "2025-01-01T00:00:02Z",
           "level": "INFO",
           "sequence": 2,
           "source": "executor",
           "message": "hello"
         }
       ]
     }'
   ```

5. **Query logs via DynamoDB Admin UI**:
   - Open `http://localhost:8001`
   - Navigate to Logs table
   - Filter by job_id = "job-001"
   - Verify logs are persisted

### Scenario 2: Control Plane Down, Agent Continues

1. **Start LocalStack**:
   ```bash
   docker-compose up -d
   ```

2. **Agent claims job and starts execution**
3. **Stop control plane** (or simulate unreachability):
   ```bash
   docker-compose stop  # or kill the control plane service
   ```
4. **Agent continues executing** (job execution should not block)
5. **Agent buffers logs locally** in `/tmp/smidr/logs/{job_id}-{build_id}.log`
6. **Agent queues log batch locally** (in memory)
7. **Restart control plane**:
   ```bash
   docker-compose up -d
   ```
8. **Agent resumes pushing logs** (with exponential backoff retry)
9. **Verify logs appear in DynamoDB**

### Scenario 3: Artifact Upload

1. **Agent completes job with artifacts**
2. **Agent requests presigned S3 URL**:
   ```bash
   curl -X POST http://localhost:8000/api/jobs/job-001/artifact-presigned-url \
     -H "X-Tenant-Id: test-tenant" \
     -H "X-Tenant-Id: test-tenant" \
     -H "Content-Type: application/json" \
     -d '{
       "artifact_name": "build-output.tar.gz",
       "artifact_type": "archive"
     }'
   ```
3. **Agent uploads to S3 via presigned URL**:
   ```bash
   curl -X PUT "<presigned-url>" \
     --data-binary @build-output.tar.gz
   ```
4. **Verify artifact in S3**:
   ```bash
   aws --endpoint-url http://localhost:4566 s3 ls s3://smidr-artifacts/
   ```

---

## Debugging Tips

### View LocalStack Logs

```bash
docker-compose logs -f localstack
```

### Query DynamoDB Tables

```bash
# List tables
aws --endpoint-url http://localhost:4566 dynamodb list-tables

# Scan Jobs table
aws --endpoint-url http://localhost:4566 dynamodb scan --table-name Jobs

# Get specific job
aws --endpoint-url http://localhost:4566 dynamodb get-item \
  --table-name Jobs \
  --key '{"job_id": {"S": "job-001"}, "created_at": {"S": "2025-01-01T00:00:00Z"}}'
```

### Query SQS Queues

```bash
# Get queue URL
aws --endpoint-url http://localhost:4566 sqs get-queue-url --queue-name JobQueue

# Peek at messages
aws --endpoint-url http://localhost:4566 sqs receive-message --queue-url <url> --max-number-of-messages 10
```

### Query S3 Buckets

```bash
# List buckets
aws --endpoint-url http://localhost:4566 s3 ls

# List objects in bucket
aws --endpoint-url http://localhost:4566 s3 ls s3://smidr-artifacts/

# Download object
aws --endpoint-url http://localhost:4566 s3 cp s3://smidr-artifacts/build-001.tar.gz .
```

### Test HMAC Authentication

```bash
# Generate HMAC-SHA256
TENANT_ID="test-tenant"
AGENT_ID="test-agent-001"
SECRET="dev-test-secret"
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
PAYLOAD="$TENANT_ID$AGENT_ID$TIMESTAMP"
SIGNATURE=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$SECRET" -binary | base64)

echo "X-Tenant-Id: $TENANT_ID"
echo "X-Agent-Id: $AGENT_ID"
echo "X-Timestamp: $TIMESTAMP"
echo "X-Signature: $SIGNATURE"
```

### Enable Debug Logging

```bash
# In agent config
logging:
  level: DEBUG

# In control plane (Node.js)
export LOG_LEVEL=debug
npm run dev
```

---

## Clean Up

### Remove LocalStack Data

```bash
docker-compose down -v  # -v removes volumes
```

### Reset LocalStack Without Restart

```bash
# All DynamoDB tables/SQS queues/S3 buckets will be recreated on next docker-compose up
docker-compose down
docker volume prune
docker-compose up -d
```

---

## Integration with CI/CD

For GitHub Actions or other CI/CD:

```yaml
# .github/workflows/test.yml
jobs:
  integration-test:
    runs-on: ubuntu-latest
    services:
      localstack:
        image: localstack/localstack:latest
        env:
          SERVICES: apigateway,lambda,dynamodb,sqs,s3
          DEBUG: 1
        ports:
          - 4566:4566
    steps:
      - uses: actions/checkout@v3
      - name: Run agent tests
        run: |
          export AWS_ENDPOINT_URL=http://localhost:4566
          go test ./agent/... -v
      - name: Run control plane tests
        run: |
          export AWS_ENDPOINT_URL=http://localhost:4566
          npm test
```

---

## Known Limitations

- **Lambda cold starts**: LocalStack Lambda is slower than AWS
- **Performance**: Not representative of production; use for correctness, not benchmarking
- **IAM policies**: LocalStack does basic validation only; AWS will be stricter
- **Quotas**: LocalStack has no quota enforcement (unlimited items, requests, etc.)

---

## Related

- [Agent Responsibility Contract](agent-contract.md)
- [Control Plane Responsibility Contract](control-plane.md)
- [LocalStack Documentation](https://docs.localstack.cloud/)
- [AWS CLI with LocalStack](https://docs.localstack.cloud/user-guide/integrations/aws-cli/)
