#!/usr/bin/env bash
# CA Management Script for Smidr Control Plane

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONTROL_PLANE_DIR="${SCRIPT_DIR}/../control-plane"
DATA_DIR="${CONTROL_PLANE_DIR}/data"
CA_DIR="${DATA_DIR}/ca"
CA_KEY="${CA_DIR}/ca.key.pem"
CA_CERT="${CA_DIR}/ca.crt.pem"

usage() {
    cat <<EOF
Usage: $(basename "$0") <command>

CA management utilities for Smidr control plane.

Commands:
    init            Initialize a new CA (if not already exists)
    info            Display CA certificate information
    verify <cert>   Verify a certificate was signed by this CA
    export          Export CA certificate for distribution
    rotate          Rotate CA (backup old, create new)
    secure          Set secure file permissions on CA files
    help            Show this help message
EOF
    exit 0
}

init_ca() {
    if [[ -f "${CA_KEY}" && -f "${CA_CERT}" ]]; then
        echo "CA already exists at ${CA_DIR}"
        exit 1
    fi

    mkdir -p "${CA_DIR}"
    
    echo "Generating CA private key..."
    openssl genrsa -out "${CA_KEY}" 2048
    
    echo "Generating self-signed CA certificate..."
    openssl req -new -x509 \
        -key "${CA_KEY}" \
        -out "${CA_CERT}" \
        -days 3650 \
        -subj "/CN=smidr-control-plane" \
        -addext "basicConstraints=critical,CA:TRUE" \
        -addext "keyUsage=critical,keyCertSign,cRLSign" \
        -sha256

    chmod 600 "${CA_KEY}"
    chmod 644 "${CA_CERT}"

    echo "✓ CA initialized successfully"
}

show_info() {
    if [[ ! -f "${CA_CERT}" ]]; then
        echo "Error: CA certificate not found"
        exit 1
    fi

    openssl x509 -in "${CA_CERT}" -noout -text
}

verify_cert() {
    openssl verify -CAfile "${CA_CERT}" "$1"
}

export_ca() {
    cat "${CA_CERT}"
}

secure_permissions() {
    chmod 700 "${CA_DIR}"
    chmod 600 "${CA_KEY}"
    chmod 644 "${CA_CERT}"
    echo "✓ Permissions secured"
}

case "${1:-help}" in
    init) init_ca ;;
    info) show_info ;;
    verify) verify_cert "$2" ;;
    export) export_ca ;;
    secure) secure_permissions ;;
    *) usage ;;
esac
