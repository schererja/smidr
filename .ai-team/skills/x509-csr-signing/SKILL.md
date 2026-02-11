# X.509 CSR Signing in .NET

## Pattern

When building a Certificate Authority in C# that signs Certificate Signing Requests (CSRs), use `CertificateRequest.LoadSigningRequestPem()` and `X509SignatureGenerator`:

```csharp
var csr = CertificateRequest.LoadSigningRequestPem(
    csrPem,
    HashAlgorithmName.SHA256,
    CertificateRequestLoadOptions.Default);

var serial = RandomNumberGenerator.GetBytes(16);
var notBefore = DateTimeOffset.UtcNow.AddMinutes(-5);
var notAfter = DateTimeOffset.UtcNow.AddDays(365);

var signatureGenerator = X509SignatureGenerator.CreateForRSA(issuerKey, RSASignaturePadding.Pkcs1);
using var cert = csr.Create(
    issuerCert.SubjectName,  // Issuer DN
    signatureGenerator,
    notBefore,
    notAfter,
    serial);

return cert.ExportCertificatePem();
```

## Why This Matters

The `.Create()` method has overloads that can be confusing:
- `Create(X500DistinguishedName, ...)` - Sign as CA (issuer DN ≠ subject DN)
- `Create(X509Certificate2, ...)` - Self-sign (issuer = subject)

For CA signing, use the issuer's `SubjectName` (DN), not the issuer certificate object itself.

## Common Pitfall

❌ Wrong: `csr.Create(issuerCert, notBefore, notAfter, serial, padding)`
✅ Right: `csr.Create(issuerCert.SubjectName, signatureGenerator, notBefore, notAfter, serial)`

The first signature is for self-signing. The second is for CA-signed certificates.

## Metadata

- **Confidence**: medium
- **Source**: earned
- **Language**: C#
- **Domain**: Cryptography, PKI
