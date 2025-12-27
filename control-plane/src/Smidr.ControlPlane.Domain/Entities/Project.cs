namespace Smidr.ControlPlane.Domain.Entities;

public class Project
{
    public Guid Id { get; set; }
    public Guid TenantId { get; set; }
    public string Name { get; set; } = string.Empty;
    public string Description { get; set; } = string.Empty;
    public DateTime CreatedAt { get; set; }
    public DateTime? UpdatedAt { get; set; }

    // Relationships
    public Tenant Tenant { get; set; } = null!;
    public ICollection<Job> Jobs { get; set; } = new List<Job>();
}
