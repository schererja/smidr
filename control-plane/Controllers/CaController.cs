using ControlPlane.Services;
using Microsoft.AspNetCore.Mvc;

namespace ControlPlane.Controllers;

[ApiController]
[Route("ca")]
public sealed class CaController : ControllerBase
{
    private readonly CaService _caService;

    public CaController(CaService caService)
    {
        _caService = caService;
    }

    [HttpGet("certificate")]
    public ActionResult<CaCertificateResponse> GetCaCertificate()
    {
        var certPem = _caService.GetCaCertificatePem();
        return Ok(new CaCertificateResponse(certPem));
    }
}

public sealed record CaCertificateResponse(string CertPem);
