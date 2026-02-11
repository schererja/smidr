#!/bin/bash

# Test agent registration flow
cd "$(dirname "$0")"

CONTROL_PLANE_URL="https://localhost:5001"
TEST_AGENT_ID="test-agent-$(date +%s)"
HOSTNAME="test-host"

echo "==> Testing registration flow..."
echo "Agent ID: $TEST_AGENT_ID"
echo "Hostname: $HOSTNAME"
echo ""

# Generate test CSR
echo "==> Generating test CSR..."
TEST_KEY="/tmp/test-agent.key"
TEST_CSR="/tmp/test-agent.csr"

openssl genrsa -out "$TEST_KEY" 2048 2>/dev/null
openssl req -new -key "$TEST_KEY" -out "$TEST_CSR" -subj "/CN=$TEST_AGENT_ID" 2>/dev/null

CSR_PEM=$(cat "$TEST_CSR" | awk '{printf "%s\\n", $0}' | sed 's/\\n$//')

echo "CSR generated."
echo ""

# Test registration endpoint
echo "==> Testing POST /api/agents/register..."
RESPONSE=$(curl -k -s -w "\n%{http_code}" -X POST "$CONTROL_PLANE_URL/api/agents/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"agentId\": \"$TEST_AGENT_ID\",
    \"hostname\": \"$HOSTNAME\",
    \"csrPem\": \"$CSR_PEM\",
    \"os\": \"linux\"
  }")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

echo "HTTP Status: $HTTP_CODE"
echo "Response: $BODY"
echo ""

if [ "$HTTP_CODE" = "200" ]; then
  echo "✅ Registration successful!"
  
  # Test if agent can be retrieved
  echo ""
  echo "==> Testing GET /api/agents/$TEST_AGENT_ID..."
  curl -k -s "$CONTROL_PLANE_URL/api/agents/$TEST_AGENT_ID" | jq '.' 2>/dev/null || echo "Failed to parse response"
  
else
  echo "❌ Registration failed with status $HTTP_CODE"
  exit 1
fi

# Cleanup
rm -f "$TEST_KEY" "$TEST_CSR"

echo ""
echo "==> Test complete"
