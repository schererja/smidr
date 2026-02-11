namespace ControlPlane.Models;

public sealed class Agent
{
    public required string Id { get; set; }
    public required string Hostname { get; set; }
    public string? Token { get; set; }
    public string? CertificatePem { get; set; }
    public DateTime? RevokedAt { get; set; }
    public DateTime RegisteredAt { get; set; }
    public DateTime? LastHeartbeatAt { get; set; }
    public HealthStatus CurrentHealth { get; set; } = HealthStatus.Learning;
    public string? OS { get; set; }
}
