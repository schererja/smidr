namespace Smidr.ControlPlane.Domain.Entities;

public class Agent
{
    public Guid Id { get; set; }
    public Guid TenantId { get; set; }
    public string Name { get; set; } = string.Empty;
    public string Hostname { get; set; } = string.Empty;
    public AgentState State { get; set; }
    public DateTime RegisteredAt { get; set; }
    public DateTime LastHeartbeatAt { get; set; }

    // Capabilities
    public string Platform { get; set; } = string.Empty;
    public string Architecture { get; set; } = string.Empty;
    public Dictionary<string, string> Labels { get; set; } = new();
    public List<string> Features { get; set; } = new();

    // Resource limits
    public long CpuMillicores { get; set; }
    public long MemoryBytes { get; set; }
    public long DiskBytes { get; set; }

    // Relationships
    public Tenant Tenant { get; set; } = null!;
    public ICollection<Job> AssignedJobs { get; set; } = new List<Job>();
}

public enum AgentState
{
    Registering,
    Idle,
    Busy,
    Offline,
    Decommissioned
}
