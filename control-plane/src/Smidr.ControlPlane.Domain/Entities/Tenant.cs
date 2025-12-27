namespace Smidr.ControlPlane.Domain.Entities;

public class Tenant
{
    public Guid Id { get; set; }
    public string Name { get; set; } = string.Empty;
    public string Slug { get; set; } = string.Empty;
    public TenantStatus Status { get; set; }
    public DateTime CreatedAt { get; set; }
    public DateTime? UpdatedAt { get; set; }

    // Relationships
    public ICollection<Agent> Agents { get; set; } = new List<Agent>();
    public ICollection<Job> Jobs { get; set; } = new List<Job>();
    public ICollection<Project> Projects { get; set; } = new List<Project>();
}

public enum TenantStatus
{
    Active,
    Suspended,
    Deleted
}
