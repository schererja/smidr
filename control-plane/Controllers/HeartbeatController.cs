using ControlPlane.Data;
using ControlPlane.Models;
using ControlPlane.Services;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;

namespace ControlPlane.Controllers;

[ApiController]
[Route("v0/agents")]
public sealed class HeartbeatController : ControllerBase
{
    private readonly ControlPlaneDbContext _db;
    private readonly IServiceScopeFactory _scopeFactory;

    public HeartbeatController(ControlPlaneDbContext db, IServiceScopeFactory scopeFactory)
    {
        _db = db;
        _scopeFactory = scopeFactory;
    }

    [HttpPost("heartbeat")]
    public async Task<IActionResult> Heartbeat(
        [FromBody] HeartbeatRequest request,
        CancellationToken cancellationToken)
    {
        var agentIdFromCert = HttpContext.Items["AgentId"] as string;
        
        if (string.IsNullOrWhiteSpace(request.AgentId))
        {
            return BadRequest("agentId is required");
        }

        if (!string.IsNullOrWhiteSpace(agentIdFromCert) && request.AgentId != agentIdFromCert)
        {
            return StatusCode(StatusCodes.Status403Forbidden, "Agent ID mismatch");
        }

        var agent = await _db.Agents
            .FirstOrDefaultAsync(a => a.Id == request.AgentId, cancellationToken);

        if (agent == null)
        {
            return NotFound($"Agent {request.AgentId} not registered");
        }

        if (agent.RevokedAt != null)
        {
            return StatusCode(StatusCodes.Status403Forbidden, "Agent certificate revoked");
        }

        var heartbeat = new Heartbeat
        {
            AgentId = request.AgentId,
            Timestamp = request.Timestamp,
            UptimeSeconds = request.UptimeSeconds,
            LoadAverage1m = request.LoadAverage1m,
            MemoryUsedPct = request.MemoryUsedPct,
            DiskUsedPct = request.DiskUsedPct,
            ProcessCount = request.ProcessCount,
            ReceivedAt = DateTime.UtcNow
        };

        _db.Heartbeats.Add(heartbeat);
        agent.LastHeartbeatAt = DateTime.UtcNow;

        await _db.SaveChangesAsync(cancellationToken);

        _ = Task.Run(async () =>
        {
            using var scope = _scopeFactory.CreateScope();
            var healthService = scope.ServiceProvider.GetRequiredService<HealthEvaluationService>();
            await healthService.EvaluateAgentHealthAsync(request.AgentId, CancellationToken.None);
        });

        return Ok();
    }
}

public sealed record HeartbeatRequest(
    string AgentId,
    DateTime Timestamp,
    double UptimeSeconds,
    double LoadAverage1m,
    double MemoryUsedPct,
    double DiskUsedPct,
    int ProcessCount);
