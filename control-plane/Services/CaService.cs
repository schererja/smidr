using System.Security.Cryptography;
using System.Security.Cryptography.X509Certificates;

namespace ControlPlane.Services;

public sealed class CaService
{
    private readonly string _caKeyPath;
    private readonly string _caCertPath;
    private readonly X509Certificate2 _issuerCert;
    private readonly RSA _issuerKey;

    public CaService(string dataDir)
    {
        var caDir = Path.Combine(dataDir, "ca");
        Directory.CreateDirectory(caDir);

        _caKeyPath = Path.Combine(caDir, "ca.key.pem");
        _caCertPath = Path.Combine(caDir, "ca.crt.pem");

        if (File.Exists(_caKeyPath) && File.Exists(_caCertPath))
        {
            (_issuerKey, _issuerCert) = LoadExisting();
        }
        else
        {
            (_issuerKey, _issuerCert) = CreateNew();
        }
    }

    public string GetCaCertificatePem()
    {
        return _issuerCert.ExportCertificatePem();
    }

    public string IssueAgentCertificate(string csrPem)
    {
        if (string.IsNullOrWhiteSpace(csrPem))
        {
            throw new ArgumentException("CSR PEM is required", nameof(csrPem));
        }

        var csr = CertificateRequest.LoadSigningRequestPem(
            csrPem,
            HashAlgorithmName.SHA256,
            CertificateRequestLoadOptions.Default);

        var serial = RandomNumberGenerator.GetBytes(16);
        var notBefore = DateTimeOffset.UtcNow.AddMinutes(-5);
        var notAfter = DateTimeOffset.UtcNow.AddDays(365);

        var signatureGenerator = X509SignatureGenerator.CreateForRSA(_issuerKey, RSASignaturePadding.Pkcs1);
        using var cert = csr.Create(
            _issuerCert.SubjectName,
            signatureGenerator,
            notBefore,
            notAfter,
            serial);

        return cert.ExportCertificatePem();
    }

    private (RSA key, X509Certificate2 cert) LoadExisting()
    {
        var keyPem = File.ReadAllText(_caKeyPath);
        var certPem = File.ReadAllText(_caCertPath);

        var key = RSA.Create();
        key.ImportFromPem(keyPem);

        var cert = X509Certificate2.CreateFromPem(certPem).CopyWithPrivateKey(key);
        return (key, cert);
    }

    private (RSA key, X509Certificate2 cert) CreateNew()
    {
        var key = RSA.Create(2048);

        var req = new CertificateRequest(
            "CN=smidr-control-plane",
            key,
            HashAlgorithmName.SHA256,
            RSASignaturePadding.Pkcs1);

        req.CertificateExtensions.Add(new X509BasicConstraintsExtension(true, false, 0, true));
        req.CertificateExtensions.Add(
            new X509KeyUsageExtension(X509KeyUsageFlags.KeyCertSign | X509KeyUsageFlags.CrlSign, true));
        req.CertificateExtensions.Add(new X509SubjectKeyIdentifierExtension(req.PublicKey, false));

        var cert = req.CreateSelfSigned(DateTimeOffset.UtcNow.AddDays(-1), DateTimeOffset.UtcNow.AddYears(10));

        File.WriteAllText(_caKeyPath, key.ExportPkcs8PrivateKeyPem());
        File.WriteAllText(_caCertPath, cert.ExportCertificatePem());

        return (key, cert);
    }
}
