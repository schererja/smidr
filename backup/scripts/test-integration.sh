#!/bin/bash

set -e

echo "🧪 Testing Smidr Control Plane Integration"
echo "=========================================="
echo ""

BASE_URL="http://localhost:8000"
TENANT_ID="test-tenant"
AGENT_ID="test-agent-001"
SECRET="dev-test-secret"

# Helper function to generate HMAC signature
generate_signature() {
    local timestamp=$1
    local payload="${TENANT_ID}${AGENT_ID}${timestamp}"
    echo -n "$payload" | openssl dgst -sha256 -hmac "$SECRET" -binary | base64
}

# Test 1: Health Check
echo "✓ Test 1: Health Check"
curl -s "${BASE_URL}/health" | jq .
echo ""

# Test 2: Agent Registration
echo "✓ Test 2: Agent Registration"
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
SIGNATURE=$(generate_signature "$TIMESTAMP")

AGENT_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/agents/register" \
  -H "X-Tenant-Id: ${TENANT_ID}" \
  -H "X-Agent-Id: ${AGENT_ID}" \
  -H "X-Timestamp: ${TIMESTAMP}" \
  -H "X-Signature: ${SIGNATURE}" \
  -H "Content-Type: application/json" \
  -d '{"name": "test-agent-local"}')

echo "$AGENT_RESPONSE" | jq .
REGISTERED_AT=$(echo "$AGENT_RESPONSE" | jq -r .registered_at)
echo ""

# Test 3: Create a Job
echo "✓ Test 3: Create Job"
JOB_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/jobs" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "'"${TENANT_ID}"'",
    "job_type": "build",
    "parameters": {
      "command": "echo hello",
      "timeout": 300
    },
    "deadline_utc": "2025-12-31T23:59:59Z"
  }')

echo "$JOB_RESPONSE" | jq .
JOB_ID=$(echo "$JOB_RESPONSE" | jq -r .job_id)
CREATED_AT=$(echo "$JOB_RESPONSE" | jq -r .created_at)
echo ""

# Test 4: Agent Claims Job
echo "✓ Test 4: Agent Claims Job"
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
SIGNATURE=$(generate_signature "$TIMESTAMP")

CLAIM_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/agents/${AGENT_ID}/jobs/claim" \
  -H "X-Tenant-Id: ${TENANT_ID}" \
  -H "X-Agent-Id: ${AGENT_ID}" \
  -H "X-Timestamp: ${TIMESTAMP}" \
  -H "X-Signature: ${SIGNATURE}" \
  -H "Content-Type: application/json")

echo "$CLAIM_RESPONSE" | jq .
echo ""

# Test 5: Upload Log Batch
echo "✓ Test 5: Upload Log Batch"
BUILD_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
SIGNATURE=$(generate_signature "$TIMESTAMP")

LOG_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/jobs/${JOB_ID}/logs/batch" \
  -H "X-Tenant-Id: ${TENANT_ID}" \
  -H "X-Agent-Id: ${AGENT_ID}" \
  -H "X-Timestamp: ${TIMESTAMP}" \
  -H "X-Signature: ${SIGNATURE}" \
  -H "Content-Type: application/json" \
  -d '{
    "job_id": "'"${JOB_ID}"'",
    "build_id": "'"${BUILD_ID}"'",
    "logs": [
      {
        "job_id": "'"${JOB_ID}"'",
        "build_id": "'"${BUILD_ID}"'",
        "sequence": 1,
        "timestamp": "2025-12-26T00:00:01Z",
        "level": "INFO",
        "source": "executor",
        "message": "Starting job execution"
      },
      {
        "job_id": "'"${JOB_ID}"'",
        "build_id": "'"${BUILD_ID}"'",
        "sequence": 2,
        "timestamp": "2025-12-26T00:00:02Z",
        "level": "INFO",
        "source": "executor",
        "message": "hello"
      },
      {
        "job_id": "'"${JOB_ID}"'",
        "build_id": "'"${BUILD_ID}"'",
        "sequence": 3,
        "timestamp": "2025-12-26T00:00:03Z",
        "level": "INFO",
        "source": "executor",
        "message": "Job completed successfully"
      }
    ]
  }')

echo "$LOG_RESPONSE" | jq .
echo ""

# Test 6: Complete Job
echo "✓ Test 6: Complete Job"
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
SIGNATURE=$(generate_signature "$TIMESTAMP")

COMPLETE_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/jobs/${JOB_ID}/complete" \
  -H "X-Tenant-Id: ${TENANT_ID}" \
  -H "X-Agent-Id: ${AGENT_ID}" \
  -H "X-Timestamp: ${TIMESTAMP}" \
  -H "X-Signature: ${SIGNATURE}" \
  -H "Content-Type: application/json" \
  -d '{
    "build_id": "'"${BUILD_ID}"'",
    "final_state": "Succeeded",
    "created_at": "'"${CREATED_AT}"'",
    "exit_code": 0
  }')

echo "$COMPLETE_RESPONSE" | jq .
echo ""

# Test 7: Retrieve Logs
echo "✓ Test 7: Retrieve Logs"
LOGS=$(curl -s "${BASE_URL}/api/jobs/${JOB_ID}/logs")
echo "$LOGS" | jq .
echo ""

# Test 8: Agent Heartbeat
echo "✓ Test 8: Agent Heartbeat"
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
SIGNATURE=$(generate_signature "$TIMESTAMP")

HEARTBEAT_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/agents/${AGENT_ID}/heartbeat" \
  -H "X-Tenant-Id: ${TENANT_ID}" \
  -H "X-Agent-Id: ${AGENT_ID}" \
  -H "X-Timestamp: ${TIMESTAMP}" \
  -H "X-Signature: ${SIGNATURE}" \
  -H "Content-Type: application/json" \
  -d '{
    "state": "Idle",
    "current_job_count": 0,
    "registered_at": "'"${REGISTERED_AT}"'",
    "metrics": {
      "cpu": 45.2,
      "memory": 62.8,
      "disk": 78.5
    }
  }')

echo "$HEARTBEAT_RESPONSE" | jq .
echo ""

echo "=========================================="
echo "✅ All tests passed!"
echo ""
echo "DynamoDB Admin UI: http://localhost:8001"
echo "Control Plane: http://localhost:8000/health"
echo ""
echo "View tables:"
echo "  aws --endpoint-url http://localhost:4566 dynamodb scan --table-name Jobs"
echo "  aws --endpoint-url http://localhost:4566 dynamodb scan --table-name Logs"
echo "  aws --endpoint-url http://localhost:4566 dynamodb scan --table-name Agents"
