---
name: "async-health-evaluation"
description: "Fire-and-forget pattern for asynchronous health evaluation after telemetry ingestion"
domain: "architecture"
confidence: "low"
source: "earned"
---

## Context
When building telemetry ingestion systems, health evaluation should not block the ingestion path. This pattern decouples telemetry storage from health computation using fire-and-forget tasks, ensuring fast response times and horizontal scalability.

## Patterns

### Fire-and-Forget Health Evaluation
After storing telemetry, trigger health evaluation asynchronously using `Task.Run()` with `CancellationToken.None`. Do not await the task — return success immediately to the client.

```csharp
// Store heartbeat synchronously
_db.Heartbeats.Add(heartbeat);
await _db.SaveChangesAsync(cancellationToken);

// Trigger evaluation asynchronously (fire-and-forget)
_ = Task.Run(() => _healthService.EvaluateAgentHealthAsync(agentId, CancellationToken.None));

return Ok();
```

### Learning Period Before Evaluation
New entities start in a "learning" state with a fixed time window (e.g., 24 hours) to establish baselines. Health evaluation during learning only computes statistics, does not raise alerts.

```csharp
var learningCutoff = DateTime.UtcNow.AddHours(-LearningPeriodHours);
if (entity.RegisteredAt > learningCutoff)
{
    // Still learning - compute baselines, don't evaluate
    await ComputeBaselinesAsync(entityId, cancellationToken);
    return;
}
```

### Statistical Baseline Computation
For each metric, compute mean, standard deviation, min, and max from samples during learning window. Store baselines in dedicated table indexed by entity + metric.

```csharp
var mean = values.Average();
var stdDev = Math.Sqrt(values.Select(v => Math.Pow(v - mean, 2)).Average());
var min = values.Min();
var max = values.Max();
```

### Anomaly Detection via Standard Deviations
Compare current value to baseline using z-score (standard deviations from mean). Threshold at 3-sigma is a good default for most metrics.

```csharp
var deviation = Math.Abs(value - baseline.Mean) / (baseline.StdDev + 0.001);
return deviation > DeviationThreshold; // e.g., 3.0
```

### Health State Transitions with Context
When health status changes, record the transition in a dedicated table with timestamp and reason. This provides audit trail and debugging context.

```csharp
_db.HealthStates.Add(new HealthState
{
    EntityId = entityId,
    Status = newStatus,
    Reason = "Anomalies detected: LoadAverage1m: 5.2 (baseline: 1.2 ± 0.8)",
    ChangedAt = DateTime.UtcNow
});
```

## Examples

### Control Plane Heartbeat Ingestion
```csharp
[HttpPost("heartbeat")]
public async Task<IActionResult> Heartbeat([FromBody] HeartbeatRequest request)
{
    var heartbeat = new Heartbeat { ... };
    _db.Heartbeats.Add(heartbeat);
    await _db.SaveChangesAsync();
    
    _ = Task.Run(() => _healthService.EvaluateAsync(request.AgentId, CancellationToken.None));
    
    return Ok();  // Fast response, evaluation happens in background
}
```

### Health Evaluation Service
```csharp
public async Task EvaluateAsync(string entityId, CancellationToken ct)
{
    // Check if still learning
    if (IsLearning(entity)) { ComputeBaselines(); return; }
    
    // Get baselines
    var baselines = await _db.Baselines.Where(b => b.EntityId == entityId).ToListAsync(ct);
    
    // Get latest telemetry
    var latest = await _db.Telemetry.OrderByDescending(t => t.Timestamp).FirstAsync(ct);
    
    // Detect anomalies
    var anomalies = baselines.Where(b => IsAnomaly(GetValue(latest, b.Metric), b)).ToList();
    
    // Update health state if changed
    var newStatus = DetermineStatus(anomalies.Count);
    if (entity.Status != newStatus) RecordTransition(entityId, newStatus, anomalies);
}
```

## Anti-Patterns
- **Awaiting health evaluation** — Blocks ingestion path, reduces throughput, creates timeout risk
- **Synchronous evaluation** — Couples telemetry storage to health computation, prevents independent scaling
- **No learning period** — Raises false positives before baselines stabilize
- **Fixed thresholds** — Cannot adapt to different workload patterns across entities
- **No transition history** — Makes debugging state changes impossible

## Benefits
- **Fast ingestion**: Telemetry API responds in <50ms regardless of evaluation complexity
- **Scalability**: Evaluation can run on separate workers/threads/processes
- **Reliability**: Evaluation failures don't block telemetry storage
- **Debuggability**: State transitions recorded with context for post-mortem analysis
- **Adaptability**: Per-entity baselines handle heterogeneous workloads
