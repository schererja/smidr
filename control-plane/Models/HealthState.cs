namespace ControlPlane.Models;

public enum HealthStatus
{
    Learning,
    Healthy,
    Degraded,
    Attention,
    Unknown
}

public sealed class HealthState
{
    public int Id { get; set; }
    public required string AgentId { get; set; }
    public HealthStatus Status { get; set; }
    public string? Reason { get; set; }
    public DateTime ChangedAt { get; set; }
}
