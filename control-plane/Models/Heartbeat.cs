namespace ControlPlane.Models;

public sealed class Heartbeat
{
    public int Id { get; set; }
    public required string AgentId { get; set; }
    public DateTime Timestamp { get; set; }
    public double UptimeSeconds { get; set; }
    public double LoadAverage1m { get; set; }
    public double MemoryUsedPct { get; set; }
    public double DiskUsedPct { get; set; }
    public int ProcessCount { get; set; }
    public DateTime ReceivedAt { get; set; }
}
