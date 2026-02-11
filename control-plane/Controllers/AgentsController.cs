using ControlPlane.Data;
using ControlPlane.Models;
using ControlPlane.Services;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;

namespace ControlPlane.Controllers;

[ApiController]
[Route("api/agents")]
public sealed class AgentsController : ControllerBase
{
    private readonly ControlPlaneDbContext _db;
    private readonly CaService _caService;

    public AgentsController(ControlPlaneDbContext db, CaService caService)
    {
        _db = db;
        _caService = caService;
    }

    [HttpGet]
    public async Task<ActionResult<IReadOnlyList<AgentListDto>>> GetAgents(CancellationToken cancellationToken)
    {
        var agents = await _db.Agents
            .OrderBy(agent => agent.RegisteredAt)
            .ToListAsync(cancellationToken);

        var agentDtos = new List<AgentListDto>();

        foreach (var agent in agents)
        {
            var latestHeartbeat = await _db.Heartbeats
                .Where(h => h.AgentId == agent.Id)
                .OrderByDescending(h => h.Timestamp)
                .FirstOrDefaultAsync(cancellationToken);

            agentDtos.Add(new AgentListDto(
                agent.Id,
                agent.Hostname,
                agent.RegisteredAt,
                agent.LastHeartbeatAt,
                agent.CurrentHealth.ToString(),
                agent.RevokedAt,
                latestHeartbeat != null ? new SignalValuesDto(
                    latestHeartbeat.UptimeSeconds,
                    latestHeartbeat.LoadAverage1m,
                    latestHeartbeat.MemoryUsedPct,
                    latestHeartbeat.DiskUsedPct,
                    latestHeartbeat.ProcessCount
                ) : null,
                agent.OS
            ));
        }

        return Ok(agentDtos);
    }

    [HttpGet("{id}")]
    public async Task<ActionResult<AgentDetailDto>> GetAgent(string id, CancellationToken cancellationToken)
    {
        var agent = await _db.Agents.FirstOrDefaultAsync(a => a.Id == id, cancellationToken);
        if (agent == null)
        {
            return NotFound();
        }

        var latestHeartbeat = await _db.Heartbeats
            .Where(h => h.AgentId == agent.Id)
            .OrderByDescending(h => h.Timestamp)
            .FirstOrDefaultAsync(cancellationToken);

        var baselines = await _db.AgentBaselines
            .Where(b => b.AgentId == agent.Id)
            .ToListAsync(cancellationToken);

        var recentHeartbeats = await _db.Heartbeats
            .Where(h => h.AgentId == agent.Id)
            .OrderByDescending(h => h.Timestamp)
            .Take(10)
            .ToListAsync(cancellationToken);

        var baselineDtos = baselines.Select(b => new BaselineDto(
            b.MetricName,
            b.Mean,
            b.StdDev,
            b.Min,
            b.Max,
            b.SampleCount
        )).ToList();

        var heartbeatDtos = recentHeartbeats.Select(h => new HeartbeatDto(
            h.Timestamp,
            h.UptimeSeconds,
            h.LoadAverage1m,
            h.MemoryUsedPct,
            h.DiskUsedPct,
            h.ProcessCount
        )).ToList();

        var agentDto = new AgentDetailDto(
            agent.Id,
            agent.Hostname,
            agent.RegisteredAt,
            agent.LastHeartbeatAt,
            agent.CurrentHealth.ToString(),
            agent.RevokedAt,
            latestHeartbeat != null ? new SignalValuesDto(
                latestHeartbeat.UptimeSeconds,
                latestHeartbeat.LoadAverage1m,
                latestHeartbeat.MemoryUsedPct,
                latestHeartbeat.DiskUsedPct,
                latestHeartbeat.ProcessCount
            ) : null,
            baselineDtos,
            heartbeatDtos,
            agent.OS
        );

        return Ok(agentDto);
    }

    [HttpPost("register")]
    public async Task<ActionResult<Agent>> RegisterAgent(
        [FromBody] RegisterAgentRequest request,
        CancellationToken cancellationToken)
    {
        if (string.IsNullOrWhiteSpace(request.AgentId))
        {
            return BadRequest("agentId is required");
        }

        if (string.IsNullOrWhiteSpace(request.Hostname))
        {
            return BadRequest("hostname is required");
        }

        if (string.IsNullOrWhiteSpace(request.CsrPem))
        {
            return BadRequest("csrPem is required");
        }

        string certPem;
        try
        {
            certPem = _caService.IssueAgentCertificate(request.CsrPem);
        }
        catch (Exception ex)
        {
            return BadRequest($"csrPem invalid: {ex.Message}");
        }

        var existing = await _db.Agents
            .FirstOrDefaultAsync(agent => agent.Id == request.AgentId, cancellationToken);

        if (existing == null)
        {
            existing = new Agent
            {
                Id = request.AgentId,
                Hostname = request.Hostname,
                Token = request.Token,
                CertificatePem = certPem,
                RevokedAt = null,
                RegisteredAt = DateTime.UtcNow,
                OS = request.OS
            };
            _db.Agents.Add(existing);
        }
        else
        {
            existing.Hostname = request.Hostname;
            existing.Token = request.Token;
            existing.CertificatePem = certPem;
            existing.RevokedAt = null;
            existing.RegisteredAt = DateTime.UtcNow;
            existing.OS = request.OS;
        }

        await _db.SaveChangesAsync(cancellationToken);

        return Ok(new RegisterAgentResponse(certPem));
    }

    [HttpPost("{id}/revoke")]
    public async Task<ActionResult<Agent>> RevokeAgent(string id, CancellationToken cancellationToken)
    {
        var agent = await _db.Agents.FirstOrDefaultAsync(a => a.Id == id, cancellationToken);
        if (agent == null)
        {
            return NotFound();
        }

        agent.RevokedAt = DateTime.UtcNow;
        await _db.SaveChangesAsync(cancellationToken);

        return Ok(agent);
    }
}

public sealed record RegisterAgentRequest(string AgentId, string Hostname, string? Token, string CsrPem, string? OS);

public sealed record RegisterAgentResponse(string CertPem);

public sealed record SignalValuesDto(
    double UptimeSeconds,
    double LoadAverage1m,
    double MemoryUsedPct,
    double DiskUsedPct,
    int ProcessCount
);

public sealed record AgentListDto(
    string Id,
    string Hostname,
    DateTime RegisteredAt,
    DateTime? LastHeartbeatAt,
    string CurrentHealth,
    DateTime? RevokedAt,
    SignalValuesDto? LatestSignals,
    string? OS
);

public sealed record BaselineDto(
    string MetricName,
    double Mean,
    double StdDev,
    double Min,
    double Max,
    int SampleCount
);

public sealed record HeartbeatDto(
    DateTime Timestamp,
    double UptimeSeconds,
    double LoadAverage1m,
    double MemoryUsedPct,
    double DiskUsedPct,
    int ProcessCount
);

public sealed record AgentDetailDto(
    string Id,
    string Hostname,
    DateTime RegisteredAt,
    DateTime? LastHeartbeatAt,
    string CurrentHealth,
    DateTime? RevokedAt,
    SignalValuesDto? LatestSignals,
    IReadOnlyList<BaselineDto> Baselines,
    IReadOnlyList<HeartbeatDto> RecentHeartbeats,
    string? OS
);
