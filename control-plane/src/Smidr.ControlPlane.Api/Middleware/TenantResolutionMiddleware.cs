using Smidr.ControlPlane.Application.Common;
using Smidr.ControlPlane.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;

namespace Smidr.ControlPlane.Api.Middleware;

public class TenantResolutionMiddleware
{
    private readonly RequestDelegate _next;
    private readonly ILogger<TenantResolutionMiddleware> _logger;

    public TenantResolutionMiddleware(
        RequestDelegate next,
        ILogger<TenantResolutionMiddleware> logger)
    {
        _next = next;
        _logger = logger;
    }

    public async Task InvokeAsync(
        HttpContext context,
        ITenantContext tenantContext,
        SmidrDbContext dbContext)
    {
        // Try to get tenant from header
        if (context.Request.Headers.TryGetValue("X-Tenant-Id", out var tenantIdHeader) &&
            Guid.TryParse(tenantIdHeader, out var tenantId))
        {
            var tenant = await dbContext.Tenants.FindAsync(tenantId);
            if (tenant != null && tenant.Status == Domain.Entities.TenantStatus.Active)
            {
                if (tenantContext is TenantContext ctx)
                {
                    ctx.SetTenant(tenant.Id, tenant.Slug);
                    _logger.LogDebug("Tenant {TenantId} resolved from header", tenantId);
                }
            }
        }
        // Try to get tenant from slug in header
        else if (context.Request.Headers.TryGetValue("X-Tenant-Slug", out var tenantSlugHeader))
        {
            var tenant = await dbContext.Tenants
                .FirstOrDefaultAsync(t => t.Slug == tenantSlugHeader && t.Status == Domain.Entities.TenantStatus.Active);
            if (tenant != null)
            {
                if (tenantContext is TenantContext ctx)
                {
                    ctx.SetTenant(tenant.Id, tenant.Slug);
                    _logger.LogDebug("Tenant {TenantSlug} resolved from header", tenantSlugHeader);
                }
            }
        }
        // For gRPC, try to get from metadata
        else if (context.Request.ContentType?.Contains("application/grpc") == true)
        {
            var metadata = context.Request.Headers;
            if (metadata.TryGetValue("x-tenant-id", out var grpcTenantId) &&
                Guid.TryParse(grpcTenantId, out var gtenantId))
            {
                var tenant = await dbContext.Tenants.FindAsync(gtenantId);
                if (tenant != null && tenant.Status == Domain.Entities.TenantStatus.Active)
                {
                    if (tenantContext is TenantContext ctx)
                    {
                        ctx.SetTenant(tenant.Id, tenant.Slug);
                        _logger.LogDebug("Tenant {TenantId} resolved from gRPC metadata", gtenantId);
                    }
                }
            }
        }

        await _next(context);
    }
}
