# mTLS Security Architecture

This document describes the cryptographic architecture and security model for Smidr's mutual TLS (mTLS) authentication system.

## Overview

Smidr uses an internal Certificate Authority (CA) to establish trust between agents and the control plane. All agent-to-control-plane communication is authenticated using X.509 client certificates over TLS 1.2+.

## Certificate Authority (CA)

### Design

- **Type**: Self-signed root CA
- **Algorithm**: RSA 2048-bit
- **Hash**: SHA-256
- **Validity**: 10 years (default)
- **Location**: `control-plane/data/ca/`
- **Files**:
  - `ca.key.pem` - Private key (600 permissions, owner-only)
  - `ca.crt.pem` - Public certificate (644 permissions, world-readable)

### Security Properties

**Protected Against**:

- Man-in-the-middle (MITM) attacks
- Agent impersonation
- Message tampering
- Eavesdropping

**Key Protection**:

- Permissions: 600 (owner read/write only)
- Storage: Local filesystem with strict access control
- Backup: Secure offline backup recommended

## Agent Certificates

### Enrollment Flow

Agent generates UUID + keypair → Creates CSR → Submits to control plane → Receives signed certificate → Uses for mTLS

### Certificate Properties

- **Subject**: CN=agent_id (UUID)
- **Algorithm**: RSA 2048-bit
- **Validity**: 365 days
- **Usage**: Client authentication

## API Endpoints

### POST /agents/register

- Enrollment endpoint (no auth required)
- Accepts CSR, returns signed certificate

### POST /v0/agents/heartbeat

- mTLS authentication required
- Validates client certificate against CA

### GET /ca/certificate

- Public endpoint to download CA certificate
- Agents can verify control plane identity

## References

- RFC 5280 - X.509 Certificate Profile
- RFC 8446 - TLS 1.3
- ASP.NET Core Certificate Authentication
