namespace Smidr.ControlPlane.Domain.Entities;

public class Job
{
    public Guid Id { get; set; }
    public Guid TenantId { get; set; }
    public Guid ProjectId { get; set; }
    public Guid? AssignedAgentId { get; set; }
    public Guid CreatedByUserId { get; set; }

    public string Name { get; set; } = string.Empty;
    public string JobType { get; set; } = string.Empty;
    public string Description { get; set; } = string.Empty;

    public JobState State { get; set; }
    public Dictionary<string, string> Parameters { get; set; } = new();
    public List<string> RequiredCapabilities { get; set; } = new();

    // Timing
    public DateTime SubmittedAt { get; set; }
    public DateTime? StartedAt { get; set; }
    public DateTime? CompletedAt { get; set; }

    // Results
    public string? ExitCode { get; set; }
    public string? ErrorMessage { get; set; }
    public int RetryCount { get; set; }
    public int MaxRetries { get; set; }

    // Relationships
    public Tenant Tenant { get; set; } = null!;
    public Project Project { get; set; } = null!;
    public Agent? AssignedAgent { get; set; }
    public ICollection<Artifact> Artifacts { get; set; } = new List<Artifact>();
}

public enum JobState
{
    Submitted,
    Queued,
    Dispatched,
    Running,
    Succeeded,
    Failed,
    Cancelled,
    TimedOut
}
