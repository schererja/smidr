#!/bin/bash

set -e

echo "Initializing LocalStack services..."

# Use LocalStack endpoint
export AWS_ENDPOINT_URL=http://localhost:4566
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test

# Wait for LocalStack to be ready
echo "Waiting for LocalStack to be ready..."
sleep 5

# Create DynamoDB Tables
echo "Creating DynamoDB tables..."

# Agents table
awslocal dynamodb create-table \
  --table-name Agents \
  --attribute-definitions AttributeName=agent_id,AttributeType=S AttributeName=registered_at,AttributeType=S \
  --key-schema AttributeName=agent_id,KeyType=HASH AttributeName=registered_at,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST \
  --stream-specification StreamEnabled=true,StreamViewType=NEW_AND_OLD_IMAGES || echo "Agents table may already exist"

# Jobs table
awslocal dynamodb create-table \
  --table-name Jobs \
  --attribute-definitions AttributeName=job_id,AttributeType=S AttributeName=created_at,AttributeType=S \
  --key-schema AttributeName=job_id,KeyType=HASH AttributeName=created_at,KeyType=RANGE \

  --billing-mode PAY_PER_REQUEST \
  --stream-specification StreamEnabled=true,StreamViewType=NEW_AND_OLD_IMAGES || echo "Jobs table may already exist"

# Logs table (index for querying logs by job_id)
awslocal dynamodb create-table \
  --table-name Logs \
  --attribute-definitions AttributeName=job_id,AttributeType=S AttributeName=sequence,AttributeType=N \
  --key-schema AttributeName=job_id,KeyType=HASH AttributeName=sequence,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST \
  --stream-specification StreamEnabled=true,StreamViewType=NEW_IMAGE || echo "Logs table may already exist"

# LogBatches table (for tracking uploaded batches, deduplication)
awslocal dynamodb create-table \
  --table-name LogBatches \
  --attribute-definitions AttributeName=job_id,AttributeType=S AttributeName=batch_id,AttributeType=S \
  --key-schema AttributeName=job_id,KeyType=HASH AttributeName=batch_id,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST || echo "LogBatches table may already exist"

# Create SQS Queue
echo "Creating SQS queue..."
awslocal sqs create-queue \
  --queue-name JobQueue \
  --attributes \
    VisibilityTimeout=300,MessageRetentionPeriod=1209600,ReceiveMessageWaitTimeSeconds=20 \
  || echo "JobQueue may already exist"

awslocal sqs create-queue \
  --queue-name JobQueue.fifo \
  --attributes \
    FifoQueue=true,ContentBasedDeduplication=true,VisibilityTimeout=300 \
  || echo "JobQueue.fifo may already exist"

# Create S3 Buckets
echo "Creating S3 buckets..."
awslocal s3 mb s3://smidr-artifacts || echo "smidr-artifacts bucket may already exist"
awslocal s3 mb s3://smidr-logs || echo "smidr-logs bucket may already exist"

# Enable versioning on S3 buckets for durability
awslocal s3api put-bucket-versioning \
  --bucket smidr-artifacts \
  --versioning-configuration Status=Enabled \
  || echo "Versioning may already be enabled"

awslocal s3api put-bucket-versioning \
  --bucket smidr-logs \
  --versioning-configuration Status=Enabled \
  || echo "Versioning may already be enabled"

echo "✓ LocalStack initialization complete!"
echo ""
echo "Services available at: http://localhost:4566"
echo "DynamoDB Admin UI: http://localhost:8001"
echo ""
echo "Endpoints for local development:"
echo "  API Gateway: http://localhost:4566"
echo "  DynamoDB: http://localhost:4566"
echo "  SQS: http://localhost:4566"
echo "  S3: http://localhost:4566"
echo ""
echo "AWS Credentials (for local testing):"
echo "  AWS_ACCESS_KEY_ID=test"
echo "  AWS_SECRET_ACCESS_KEY=test"
echo "  AWS_REGION=us-east-1"
