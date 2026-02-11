#!/usr/bin/env bash
# Quick test script to verify mTLS enrollment works

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMP_DIR=$(mktemp -d)
trap "rm -rf ${TEMP_DIR}" EXIT

echo "=== Testing mTLS Enrollment Flow ==="
echo ""

# Generate test agent credentials
echo "1. Generating test agent UUID and keypair..."
TEST_UUID=$(uuidgen | tr '[:upper:]' '[:lower:]')
openssl genrsa -out "${TEMP_DIR}/agent.key.pem" 2048 2>/dev/null
echo "   UUID: ${TEST_UUID}"

# Create CSR
echo "2. Creating CSR..."
openssl req -new \
    -key "${TEMP_DIR}/agent.key.pem" \
    -out "${TEMP_DIR}/agent.csr.pem" \
    -subj "/CN=${TEST_UUID}" 2>/dev/null

# Show CSR
echo "3. CSR created:"
openssl req -in "${TEMP_DIR}/agent.csr.pem" -noout -subject

echo ""
echo "✓ Test agent credentials generated successfully"
echo ""
echo "To test enrollment, start control plane and run:"
echo ""
echo "  CSR_PEM=\$(cat ${TEMP_DIR}/agent.csr.pem)"
echo "  curl -X POST http://localhost:5000/agents/register \\"
echo "    -H 'Content-Type: application/json' \\"
echo "    -d '{\"agentId\":\"${TEST_UUID}\",\"hostname\":\"test-agent\",\"csrPem\":\"'\"\${CSR_PEM}\"'\"}'"
echo ""
