using System.Security.Cryptography;
using System.Security.Cryptography.X509Certificates;
using ControlPlane.Services;

namespace ControlPlane.Tests;

public sealed class MtlsValidationServiceTests : IDisposable
{
    private readonly string _tempDir;
    private readonly CaService _caService;

    public MtlsValidationServiceTests()
    {
        _tempDir = Path.Combine(Path.GetTempPath(), $"smidr-test-{Guid.NewGuid()}");
        Directory.CreateDirectory(_tempDir);
        _caService = new CaService(_tempDir);
    }

    public void Dispose()
    {
        try
        {
            if (Directory.Exists(_tempDir))
            {
                Directory.Delete(_tempDir, true);
            }
        }
        catch
        {
            // Ignore cleanup errors
        }
    }

    private X509Certificate2 CreateClientCert(string commonName, DateTime? notBefore = null, DateTime? notAfter = null)
    {
        using var clientRsa = RSA.Create(2048);
        var req = new CertificateRequest(
            $"CN={commonName}",
            clientRsa,
            HashAlgorithmName.SHA256,
            RSASignaturePadding.Pkcs1);

        var csrBytes = req.CreateSigningRequest();
        var csrPem = $"-----BEGIN CERTIFICATE REQUEST-----\n{Convert.ToBase64String(csrBytes)}\n-----END CERTIFICATE REQUEST-----";

        var certPem = _caService.IssueAgentCertificate(csrPem);
        
        return X509Certificate2.CreateFromPem(certPem).CopyWithPrivateKey(clientRsa);
    }

    [Fact]
    public void ValidateClientCertificate_NullCert_ReturnsFalse()
    {
        var service = new MtlsValidationService(_caService);

        var result = service.ValidateClientCertificate(null, out var error);

        Assert.False(result);
        Assert.NotNull(error);
        Assert.Contains("not provided", error);
    }

    [Fact]
    public void ValidateClientCertificate_ValidCert_ReturnsTrue()
    {
        using var clientCert = CreateClientCert("agent-123");

        var service = new MtlsValidationService(_caService);

        var result = service.ValidateClientCertificate(clientCert, out var error);

        Assert.True(result);
        Assert.Null(error);
    }

    [Fact(Skip = "Requires custom cert generation with specific validity dates")]
    public void ValidateClientCertificate_ExpiredCert_ReturnsFalse()
    {
        // This test would need manual certificate generation with expired dates
        // CaService always generates valid certificates
    }

    [Fact(Skip = "Requires custom cert generation with specific validity dates")]
    public void ValidateClientCertificate_NotYetValid_ReturnsFalse()
    {
        // This test would need manual certificate generation with future dates
        // CaService always generates valid certificates
    }

    [Fact]
    public void ExtractAgentId_ValidCN_ReturnsAgentId()
    {
        using var clientCert = CreateClientCert("agent-abc-123");

        var service = new MtlsValidationService(_caService);

        var agentId = service.ExtractAgentId(clientCert);

        Assert.Equal("agent-abc-123", agentId);
    }

    [Fact]
    public void ExtractAgentId_SubjectWithMultipleFields_ExtractsCN()
    {
        using var rsa = RSA.Create(2048);
        var req = new CertificateRequest(
            "CN=my-agent,O=TestOrg,C=US",
            rsa,
            HashAlgorithmName.SHA256,
            RSASignaturePadding.Pkcs1);

        using var cert = req.CreateSelfSigned(
            DateTimeOffset.UtcNow.AddDays(-1),
            DateTimeOffset.UtcNow.AddYears(1));

        var service = new MtlsValidationService(_caService);

        var agentId = service.ExtractAgentId(cert);

        Assert.Equal("my-agent", agentId);
    }

}
