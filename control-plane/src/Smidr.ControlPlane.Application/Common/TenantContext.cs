namespace Smidr.ControlPlane.Application.Common;

public interface ITenantContext
{
    Guid TenantId { get; }
    string TenantSlug { get; }
    bool IsSet { get; }
}

public class TenantContext : ITenantContext
{
    public Guid TenantId { get; private set; }
    public string TenantSlug { get; private set; } = string.Empty;
    public bool IsSet { get; private set; }

    public void SetTenant(Guid tenantId, string tenantSlug)
    {
        TenantId = tenantId;
        TenantSlug = tenantSlug;
        IsSet = true;
    }
}
