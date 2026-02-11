namespace ControlPlane.Models;

public sealed class AgentBaseline
{
    public int Id { get; set; }
    public required string AgentId { get; set; }
    public required string MetricName { get; set; }
    public double Mean { get; set; }
    public double StdDev { get; set; }
    public double Min { get; set; }
    public double Max { get; set; }
    public int SampleCount { get; set; }
    public DateTime CreatedAt { get; set; }
    public DateTime UpdatedAt { get; set; }
}
