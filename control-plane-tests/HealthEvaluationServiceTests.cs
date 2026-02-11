using ControlPlane.Data;
using ControlPlane.Models;
using ControlPlane.Services;
using Microsoft.EntityFrameworkCore;

namespace ControlPlane.Tests;

public sealed class HealthEvaluationServiceTests
{
    private static ControlPlaneDbContext CreateInMemoryDb()
    {
        var options = new DbContextOptionsBuilder<ControlPlaneDbContext>()
            .UseInMemoryDatabase(databaseName: Guid.NewGuid().ToString())
            .Options;
        return new ControlPlaneDbContext(options);
    }

    [Fact]
    public async Task EvaluateAgentHealth_NewAgent_SetsLearningStatus()
    {
        using var db = CreateInMemoryDb();
        var service = new HealthEvaluationService(db);
        
        var agent = new Agent
        {
            Id = "test-agent",
            Hostname = "test-host",
            RegisteredAt = DateTime.UtcNow.AddMinutes(-10),
            CurrentHealth = HealthStatus.Unknown
        };
        db.Agents.Add(agent);
        await db.SaveChangesAsync();

        await service.EvaluateAgentHealthAsync("test-agent", CancellationToken.None);

        var updated = await db.Agents.FirstAsync(a => a.Id == "test-agent");
        Assert.Equal(HealthStatus.Learning, updated.CurrentHealth);
        
        var healthChange = await db.HealthStates.FirstOrDefaultAsync(h => h.AgentId == "test-agent");
        Assert.NotNull(healthChange);
        Assert.Equal(HealthStatus.Learning, healthChange.Status);
        Assert.Contains("learning period", healthChange.Reason);
    }

    [Fact]
    public async Task EvaluateAgentHealth_NoHeartbeats_SetsUnknown()
    {
        using var db = CreateInMemoryDb();
        var service = new HealthEvaluationService(db);
        
        var agent = new Agent
        {
            Id = "test-agent",
            Hostname = "test-host",
            RegisteredAt = DateTime.UtcNow.AddHours(-30),
            CurrentHealth = HealthStatus.Learning
        };
        db.Agents.Add(agent);
        await db.SaveChangesAsync();

        await service.EvaluateAgentHealthAsync("test-agent", CancellationToken.None);

        var updated = await db.Agents.FirstAsync(a => a.Id == "test-agent");
        Assert.Equal(HealthStatus.Unknown, updated.CurrentHealth);
        
        var healthChange = await db.HealthStates.FirstOrDefaultAsync(h => h.AgentId == "test-agent");
        Assert.NotNull(healthChange);
        Assert.Contains("No heartbeats", healthChange.Reason);
    }

    [Fact]
    public async Task EvaluateAgentHealth_WithBaselinesNormal_SetsHealthy()
    {
        using var db = CreateInMemoryDb();
        var service = new HealthEvaluationService(db);
        
        var agent = new Agent
        {
            Id = "test-agent",
            Hostname = "test-host",
            RegisteredAt = DateTime.UtcNow.AddHours(-30),
            CurrentHealth = HealthStatus.Learning
        };
        db.Agents.Add(agent);

        var baseline = new AgentBaseline
        {
            AgentId = "test-agent",
            MetricName = "LoadAverage1m",
            Mean = 1.0,
            StdDev = 0.2,
            Min = 0.5,
            Max = 1.5,
            SampleCount = 100,
            CreatedAt = DateTime.UtcNow,
            UpdatedAt = DateTime.UtcNow
        };
        db.AgentBaselines.Add(baseline);

        var heartbeat = new Heartbeat
        {
            AgentId = "test-agent",
            Timestamp = DateTime.UtcNow,
            UptimeSeconds = 1000,
            LoadAverage1m = 1.1,
            MemoryUsedPct = 50,
            DiskUsedPct = 30,
            ProcessCount = 100,
            ReceivedAt = DateTime.UtcNow
        };
        db.Heartbeats.Add(heartbeat);
        await db.SaveChangesAsync();

        await service.EvaluateAgentHealthAsync("test-agent", CancellationToken.None);

        var updated = await db.Agents.FirstAsync(a => a.Id == "test-agent");
        Assert.Equal(HealthStatus.Healthy, updated.CurrentHealth);
    }

    [Fact]
    public async Task EvaluateAgentHealth_OneAnomaly_SetsDegraded()
    {
        using var db = CreateInMemoryDb();
        var service = new HealthEvaluationService(db);
        
        var agent = new Agent
        {
            Id = "test-agent",
            Hostname = "test-host",
            RegisteredAt = DateTime.UtcNow.AddHours(-30),
            CurrentHealth = HealthStatus.Healthy
        };
        db.Agents.Add(agent);

        db.AgentBaselines.Add(new AgentBaseline
        {
            AgentId = "test-agent",
            MetricName = "LoadAverage1m",
            Mean = 1.0,
            StdDev = 0.2,
            Min = 0.5,
            Max = 1.5,
            SampleCount = 100,
            CreatedAt = DateTime.UtcNow,
            UpdatedAt = DateTime.UtcNow
        });

        var heartbeat = new Heartbeat
        {
            AgentId = "test-agent",
            Timestamp = DateTime.UtcNow,
            UptimeSeconds = 1000,
            LoadAverage1m = 2.5,
            MemoryUsedPct = 50,
            DiskUsedPct = 30,
            ProcessCount = 100,
            ReceivedAt = DateTime.UtcNow
        };
        db.Heartbeats.Add(heartbeat);
        await db.SaveChangesAsync();

        await service.EvaluateAgentHealthAsync("test-agent", CancellationToken.None);

        var updated = await db.Agents.FirstAsync(a => a.Id == "test-agent");
        Assert.Equal(HealthStatus.Degraded, updated.CurrentHealth);
        
        var healthChange = await db.HealthStates.OrderByDescending(h => h.ChangedAt)
            .FirstAsync(h => h.AgentId == "test-agent");
        Assert.Contains("LoadAverage1m", healthChange.Reason);
    }

    [Fact]
    public async Task EvaluateAgentHealth_MultipleAnomalies_SetsAttention()
    {
        using var db = CreateInMemoryDb();
        var service = new HealthEvaluationService(db);
        
        var agent = new Agent
        {
            Id = "test-agent",
            Hostname = "test-host",
            RegisteredAt = DateTime.UtcNow.AddHours(-30),
            CurrentHealth = HealthStatus.Healthy
        };
        db.Agents.Add(agent);

        db.AgentBaselines.AddRange(
            new AgentBaseline
            {
                AgentId = "test-agent",
                MetricName = "LoadAverage1m",
                Mean = 1.0,
                StdDev = 0.2,
                Min = 0.5,
                Max = 1.5,
                SampleCount = 100,
                CreatedAt = DateTime.UtcNow,
                UpdatedAt = DateTime.UtcNow
            },
            new AgentBaseline
            {
                AgentId = "test-agent",
                MetricName = "MemoryUsedPct",
                Mean = 50.0,
                StdDev = 5.0,
                Min = 40,
                Max = 60,
                SampleCount = 100,
                CreatedAt = DateTime.UtcNow,
                UpdatedAt = DateTime.UtcNow
            }
        );

        var heartbeat = new Heartbeat
        {
            AgentId = "test-agent",
            Timestamp = DateTime.UtcNow,
            UptimeSeconds = 1000,
            LoadAverage1m = 2.5,
            MemoryUsedPct = 80,
            DiskUsedPct = 30,
            ProcessCount = 100,
            ReceivedAt = DateTime.UtcNow
        };
        db.Heartbeats.Add(heartbeat);
        await db.SaveChangesAsync();

        await service.EvaluateAgentHealthAsync("test-agent", CancellationToken.None);

        var updated = await db.Agents.FirstAsync(a => a.Id == "test-agent");
        Assert.Equal(HealthStatus.Attention, updated.CurrentHealth);
        
        var healthChange = await db.HealthStates.OrderByDescending(h => h.ChangedAt)
            .FirstAsync(h => h.AgentId == "test-agent");
        Assert.Contains("LoadAverage1m", healthChange.Reason);
        Assert.Contains("MemoryUsedPct", healthChange.Reason);
    }

    [Fact]
    public async Task EvaluateAgentHealth_NonexistentAgent_NoError()
    {
        using var db = CreateInMemoryDb();
        var service = new HealthEvaluationService(db);

        await service.EvaluateAgentHealthAsync("nonexistent", CancellationToken.None);

        Assert.Empty(db.HealthStates);
    }

    [Fact]
    public async Task EvaluateAgentHealth_ComputesBaselinesInLearningPeriod()
    {
        using var db = CreateInMemoryDb();
        var service = new HealthEvaluationService(db);
        
        var agent = new Agent
        {
            Id = "test-agent",
            Hostname = "test-host",
            RegisteredAt = DateTime.UtcNow.AddHours(-20),
            CurrentHealth = HealthStatus.Learning
        };
        db.Agents.Add(agent);

        for (int i = 0; i < 15; i++)
        {
            db.Heartbeats.Add(new Heartbeat
            {
                AgentId = "test-agent",
                Timestamp = DateTime.UtcNow.AddHours(-19 + i),
                UptimeSeconds = 1000 + i * 100,
                LoadAverage1m = 1.0 + (i * 0.1),
                MemoryUsedPct = 50 + i,
                DiskUsedPct = 30,
                ProcessCount = 100 + i,
                ReceivedAt = DateTime.UtcNow.AddHours(-19 + i)
            });
        }
        await db.SaveChangesAsync();

        await service.EvaluateAgentHealthAsync("test-agent", CancellationToken.None);

        var baselines = await db.AgentBaselines.Where(b => b.AgentId == "test-agent").ToListAsync();
        Assert.NotEmpty(baselines);
        Assert.Contains(baselines, b => b.MetricName == "LoadAverage1m");
        Assert.Contains(baselines, b => b.MetricName == "MemoryUsedPct");
        
        var loadBaseline = baselines.First(b => b.MetricName == "LoadAverage1m");
        Assert.True(loadBaseline.Mean > 0);
        Assert.True(loadBaseline.StdDev >= 0);
        Assert.Equal(15, loadBaseline.SampleCount);
    }
}
