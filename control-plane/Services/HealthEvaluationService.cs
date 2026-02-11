using ControlPlane.Data;
using ControlPlane.Models;
using Microsoft.EntityFrameworkCore;

namespace ControlPlane.Services;

public sealed class HealthEvaluationService
{
    private readonly ControlPlaneDbContext _db;
    private const int LearningPeriodHours = 24;
    private const double DeviationThreshold = 3.0;

    public HealthEvaluationService(ControlPlaneDbContext db)
    {
        _db = db;
    }

    public async Task EvaluateAgentHealthAsync(string agentId, CancellationToken cancellationToken)
    {
        var agent = await _db.Agents.FirstOrDefaultAsync(a => a.Id == agentId, cancellationToken);
        if (agent == null) return;

        var learningCutoff = DateTime.UtcNow.AddHours(-LearningPeriodHours);
        
        if (agent.RegisteredAt > learningCutoff)
        {
            if (agent.CurrentHealth != HealthStatus.Learning)
            {
                agent.CurrentHealth = HealthStatus.Learning;
                await RecordHealthChange(agentId, HealthStatus.Learning, "Agent still in learning period", cancellationToken);
            }
            await ComputeBaselinesAsync(agentId, cancellationToken);
            return;
        }

        var baselines = await _db.AgentBaselines
            .Where(b => b.AgentId == agentId)
            .ToListAsync(cancellationToken);

        if (baselines.Count == 0)
        {
            await ComputeBaselinesAsync(agentId, cancellationToken);
            baselines = await _db.AgentBaselines
                .Where(b => b.AgentId == agentId)
                .ToListAsync(cancellationToken);
        }

        var recentHeartbeat = await _db.Heartbeats
            .Where(h => h.AgentId == agentId)
            .OrderByDescending(h => h.Timestamp)
            .FirstOrDefaultAsync(cancellationToken);

        if (recentHeartbeat == null)
        {
            if (agent.CurrentHealth != HealthStatus.Unknown)
            {
                agent.CurrentHealth = HealthStatus.Unknown;
                await RecordHealthChange(agentId, HealthStatus.Unknown, "No heartbeats received", cancellationToken);
            }
            return;
        }

        var anomalies = new List<string>();
        foreach (var baseline in baselines)
        {
            var currentValue = GetMetricValue(recentHeartbeat, baseline.MetricName);
            if (IsAnomaly(currentValue, baseline))
            {
                anomalies.Add($"{baseline.MetricName}: {currentValue:F2} (baseline: {baseline.Mean:F2} ± {baseline.StdDev:F2})");
            }
        }

        var newStatus = anomalies.Count switch
        {
            0 => HealthStatus.Healthy,
            1 => HealthStatus.Degraded,
            _ => HealthStatus.Attention
        };

        if (agent.CurrentHealth != newStatus)
        {
            var reason = anomalies.Count > 0 
                ? $"Anomalies detected: {string.Join(", ", anomalies)}"
                : "All metrics within normal range";
            
            agent.CurrentHealth = newStatus;
            await RecordHealthChange(agentId, newStatus, reason, cancellationToken);
        }

        await _db.SaveChangesAsync(cancellationToken);
    }

    private async Task ComputeBaselinesAsync(string agentId, CancellationToken cancellationToken)
    {
        var learningStart = DateTime.UtcNow.AddHours(-LearningPeriodHours);
        var heartbeats = await _db.Heartbeats
            .Where(h => h.AgentId == agentId && h.Timestamp >= learningStart)
            .OrderBy(h => h.Timestamp)
            .ToListAsync(cancellationToken);

        if (heartbeats.Count < 10) return;

        var metrics = new[] { "LoadAverage1m", "MemoryUsedPct", "DiskUsedPct", "ProcessCount" };
        
        foreach (var metric in metrics)
        {
            var values = heartbeats.Select(h => GetMetricValue(h, metric)).ToList();
            
            var mean = values.Average();
            var stdDev = Math.Sqrt(values.Select(v => Math.Pow(v - mean, 2)).Average());
            var min = values.Min();
            var max = values.Max();

            var existing = await _db.AgentBaselines
                .FirstOrDefaultAsync(b => b.AgentId == agentId && b.MetricName == metric, cancellationToken);

            if (existing == null)
            {
                _db.AgentBaselines.Add(new AgentBaseline
                {
                    AgentId = agentId,
                    MetricName = metric,
                    Mean = mean,
                    StdDev = stdDev,
                    Min = min,
                    Max = max,
                    SampleCount = values.Count,
                    CreatedAt = DateTime.UtcNow,
                    UpdatedAt = DateTime.UtcNow
                });
            }
            else
            {
                existing.Mean = mean;
                existing.StdDev = stdDev;
                existing.Min = min;
                existing.Max = max;
                existing.SampleCount = values.Count;
                existing.UpdatedAt = DateTime.UtcNow;
            }
        }

        await _db.SaveChangesAsync(cancellationToken);
    }

    private static double GetMetricValue(Heartbeat heartbeat, string metricName) => metricName switch
    {
        "LoadAverage1m" => heartbeat.LoadAverage1m,
        "MemoryUsedPct" => heartbeat.MemoryUsedPct,
        "DiskUsedPct" => heartbeat.DiskUsedPct,
        "ProcessCount" => heartbeat.ProcessCount,
        _ => 0
    };

    private static bool IsAnomaly(double value, AgentBaseline baseline)
    {
        var deviation = Math.Abs(value - baseline.Mean) / (baseline.StdDev + 0.001);
        return deviation > DeviationThreshold;
    }

    private async Task RecordHealthChange(string agentId, HealthStatus status, string reason, CancellationToken cancellationToken)
    {
        _db.HealthStates.Add(new HealthState
        {
            AgentId = agentId,
            Status = status,
            Reason = reason,
            ChangedAt = DateTime.UtcNow
        });
        await _db.SaveChangesAsync(cancellationToken);
    }
}
