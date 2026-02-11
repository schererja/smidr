using System.Security.Cryptography.X509Certificates;
using ControlPlane.Services;

namespace ControlPlane.Middleware;

public sealed class MtlsAuthenticationMiddleware
{
    private readonly RequestDelegate _next;
    private static readonly HashSet<string> _publicEndpoints = new(StringComparer.OrdinalIgnoreCase)
    {
        "/api/agents/register",
        "/api/agents",
        "/api/ca/certificate",
        "/swagger",
        "/swagger/index.html",
        "/swagger/v1/swagger.json",
    };

    public MtlsAuthenticationMiddleware(RequestDelegate next)
    {
        _next = next;
    }

    public async Task InvokeAsync(HttpContext context, MtlsValidationService mtlsValidation)
    {
        var path = context.Request.Path.Value ?? "";

        var isPublic = _publicEndpoints.Any(p => path.StartsWith(p, StringComparison.OrdinalIgnoreCase));
        if (isPublic)
        {
            await _next(context);
            return;
        }

        var clientCert = context.Connection.ClientCertificate;
        if (clientCert == null)
        {
            context.Response.StatusCode = StatusCodes.Status401Unauthorized;
            await context.Response.WriteAsync("Client certificate required");
            return;
        }

        if (!mtlsValidation.ValidateClientCertificate(clientCert, out var errorMessage))
        {
            context.Response.StatusCode = StatusCodes.Status403Forbidden;
            await context.Response.WriteAsync($"Certificate validation failed: {errorMessage}");
            return;
        }

        var agentId = mtlsValidation.ExtractAgentId(clientCert);
        if (!string.IsNullOrWhiteSpace(agentId))
        {
            context.Items["AgentId"] = agentId;
        }

        await _next(context);
    }
}
