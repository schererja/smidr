---
name: "go-mtls-client-config"
description: "Configuring mTLS client authentication in Go with proper certificate handling"
domain: "security"
confidence: "medium"
source: "earned"
---

## Context
When implementing mTLS (mutual TLS) client authentication in Go, the client must load its certificate and private key, then configure the TLS client to present them during the handshake. The server must also be configured to request client certificates during the TLS handshake for the client to send them.

## Patterns

### Basic mTLS Client Configuration

Load the client certificate and key, then attach to the HTTP client:

```go
import (
    "crypto/tls"
    "crypto/x509"
    "net/http"
    "os"
)

func newMTLSClient(certPath, keyPath, caCertPath string) (*http.Client, error) {
    // Load client certificate and private key
    cert, err := tls.LoadX509KeyPair(certPath, keyPath)
    if err != nil {
        return nil, fmt.Errorf("load client certificate: %w", err)
    }
    
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        MinVersion:   tls.VersionTLS12,
    }
    
    // Optional: Load CA cert for server verification
    if caCertPath != "" {
        caCert, err := os.ReadFile(caCertPath)
        if err != nil {
            return nil, fmt.Errorf("read ca cert: %w", err)
        }
        
        caCertPool := x509.NewCertPool()
        if !caCertPool.AppendCertsFromPEM(caCert) {
            return nil, errors.New("failed to parse ca certificate")
        }
        
        tlsConfig.RootCAs = caCertPool
    }
    
    return &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: tlsConfig,
        },
    }, nil
}
```

**Key points:**
- Use `tls.LoadX509KeyPair()` to load both cert and key in one call
- Set `Certificates` field in `tls.Config` (plural, it's a slice)
- Attach TLS config to the HTTP transport, not the client directly
- Set `MinVersion` to enforce minimum TLS version

### Server Verification with Custom CA

For self-signed or internal CA certificates:

```go
caCert, err := os.ReadFile(caCertPath)
if err != nil {
    return nil, fmt.Errorf("read ca cert: %w", err)
}

caCertPool := x509.NewCertPool()
if !caCertPool.AppendCertsFromPEM(caCert) {
    return nil, errors.New("failed to parse ca certificate")
}

tlsConfig.RootCAs = caCertPool  // Server cert must be signed by this CA
```

**Why:** System CA pool doesn't include self-signed or internal CA certificates. Creating a custom `RootCAs` pool allows the client to trust those certificates.

### Development Mode: Skip Verification

For localhost development with self-signed certificates:

```go
if strings.Contains(serverURL, "localhost") || strings.Contains(serverURL, "127.0.0.1") {
    if caCertPath == "" {  // Only if no CA cert provided
        tlsConfig.InsecureSkipVerify = true
    }
}
```

**Why:** Self-signed certificates in development won't be trusted by default. `InsecureSkipVerify` bypasses certificate validation.

⚠️ **Warning:** NEVER use `InsecureSkipVerify` in production. Only use for localhost/127.0.0.1 development.

### Common Pitfall: Server Must Request Client Certificate

The client certificate will ONLY be sent if the server requests it during TLS handshake:

**Server-side requirement (Kestrel/C#):**
```csharp
// ❌ Bad - server doesn't request cert
ClientCertificateMode = ClientCertificateMode.AllowCertificate

// ✅ Good - server requests cert during handshake
ClientCertificateMode = ClientCertificateMode.RequireCertificate
```

**Symptoms of misconfiguration:**
- Client code looks correct but authentication fails
- Server logs show "client certificate not provided" or 401 errors
- `openssl s_client -connect server:port` doesn't show "Acceptable client certificate CA names"

**Testing if server requests certificate:**
```bash
echo | openssl s_client -connect localhost:5001 2>&1 | grep -i "Acceptable client certificate"
```

If this returns nothing, the server isn't requesting client certificates.

### Error Handling Best Practices

Always wrap certificate loading errors with context:

```go
cert, err := tls.LoadX509KeyPair(certPath, keyPath)
if err != nil {
    return nil, fmt.Errorf("load client certificate (cert=%s, key=%s): %w", certPath, keyPath, err)
}
```

Common errors:
- File not found → Check paths in config
- Permission denied → Check file permissions (should be 0600 or 0400)
- Parse error → Cert and key must match, both must be PEM format
- "tls: private key does not match public key" → Wrong cert/key pair

## Anti-Patterns

- **Setting `Certificate` instead of `Certificates`** — The field name is plural; must be a slice
- **Attaching TLS config directly to `http.Client`** — Must go through `Transport.TLSClientConfig`
- **Using `InsecureSkipVerify` without checking for localhost** — Dangerous in production
- **Not handling certificate file errors** — Always wrap errors with context about which file failed
- **Assuming client cert will be sent automatically** — Server must request it during handshake

## When to Apply

Use mTLS client configuration when:
- API requires client certificate authentication
- Service-to-service communication needs mutual authentication
- Zero-trust architecture requires identity verification on both sides
- Compliance requires certificate-based authentication

Don't use mTLS when:
- API tokens or OAuth are sufficient
- Certificate distribution/rotation is too complex
- Public API needs simple authentication

## Troubleshooting Checklist

1. ✅ Client certificate and key files exist and are readable
2. ✅ Files are in PEM format (not DER or other formats)
3. ✅ Certificate and key are a matching pair
4. ✅ Certificate is signed by the CA the server trusts
5. ✅ Certificate is not expired
6. ✅ Server is configured to REQUEST client certificates (most common issue)
7. ✅ TLS config is attached to `Transport`, not to `Client` directly

## References

- Implementation: `agent/internal/heartbeat/client.go` (`buildTLSConfig` function)
- Go crypto/tls docs: https://pkg.go.dev/crypto/tls
- Testing: `openssl s_client -connect host:port -cert cert.pem -key key.pem`
