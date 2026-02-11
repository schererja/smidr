using System.Security.Cryptography.X509Certificates;

namespace ControlPlane.Services;

public sealed class MtlsValidationService
{
    private readonly CaService _caService;

    public MtlsValidationService(CaService caService)
    {
        _caService = caService;
    }

    public bool ValidateClientCertificate(X509Certificate2? clientCert, out string? errorMessage)
    {
        errorMessage = null;

        if (clientCert == null)
        {
            errorMessage = "Client certificate not provided";
            return false;
        }

        var caCertPem = _caService.GetCaCertificatePem();
        using var caCert = X509Certificate2.CreateFromPem(caCertPem);

        var chain = new X509Chain
        {
            ChainPolicy =
            {
                RevocationMode = X509RevocationMode.NoCheck,
                VerificationFlags = X509VerificationFlags.AllowUnknownCertificateAuthority
            }
        };

        chain.ChainPolicy.ExtraStore.Add(caCert);
        chain.ChainPolicy.TrustMode = X509ChainTrustMode.CustomRootTrust;
        chain.ChainPolicy.CustomTrustStore.Add(caCert);

        if (!chain.Build(clientCert))
        {
            var errors = string.Join("; ", chain.ChainStatus.Select(s => s.StatusInformation));
            errorMessage = $"Certificate chain validation failed: {errors}";
            return false;
        }

        var now = DateTime.UtcNow;
        if (clientCert.NotBefore > now || clientCert.NotAfter < now)
        {
            errorMessage = $"Certificate expired or not yet valid (valid: {clientCert.NotBefore} - {clientCert.NotAfter})";
            return false;
        }

        return true;
    }

    public string? ExtractAgentId(X509Certificate2 clientCert)
    {
        return clientCert.Subject.Split(',')
            .Select(s => s.Trim())
            .FirstOrDefault(s => s.StartsWith("CN=", StringComparison.OrdinalIgnoreCase))
            ?.Substring(3);
    }
}
