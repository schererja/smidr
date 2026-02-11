using System.Security.Cryptography;
using System.Security.Cryptography.X509Certificates;
using ControlPlane.Controllers;
using ControlPlane.Data;
using ControlPlane.Models;
using ControlPlane.Services;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;

namespace ControlPlane.Tests;

public sealed class AgentsControllerTests : IDisposable
{
    private readonly string _tempDir;
    private readonly CaService _caService;

    public AgentsControllerTests()
    {
        _tempDir = Path.Combine(Path.GetTempPath(), $"smidr-test-{Guid.NewGuid()}");
        Directory.CreateDirectory(_tempDir);
        _caService = new CaService(_tempDir);
    }

    public void Dispose()
    {
        try
        {
            if (Directory.Exists(_tempDir))
            {
                Directory.Delete(_tempDir, true);
            }
        }
        catch
        {
            // Ignore cleanup errors
        }
    }

    private static ControlPlaneDbContext CreateInMemoryDb()
    {
        var options = new DbContextOptionsBuilder<ControlPlaneDbContext>()
            .UseInMemoryDatabase(databaseName: Guid.NewGuid().ToString())
            .Options;
        return new ControlPlaneDbContext(options);
    }

    private static string GenerateTestCsr()
    {
        using var rsa = RSA.Create(2048);
        var req = new CertificateRequest(
            "CN=test-agent",
            rsa,
            HashAlgorithmName.SHA256,
            RSASignaturePadding.Pkcs1);

        var csrBytes = req.CreateSigningRequest();
        return $"-----BEGIN CERTIFICATE REQUEST-----\n{Convert.ToBase64String(csrBytes)}\n-----END CERTIFICATE REQUEST-----";
    }

    [Fact]
    public async Task GetAgents_EmptyDatabase_ReturnsEmptyList()
    {
        using var db = CreateInMemoryDb();
        var controller = new AgentsController(db, _caService);

        var result = await controller.GetAgents(CancellationToken.None);

        var okResult = Assert.IsType<OkObjectResult>(result.Result);
        var agents = Assert.IsAssignableFrom<IReadOnlyList<AgentListDto>>(okResult.Value);
        Assert.Empty(agents);
    }

    [Fact]
    public async Task GetAgents_WithAgents_ReturnsAgentList()
    {
        using var db = CreateInMemoryDb();
        db.Agents.Add(new Agent
        {
            Id = "agent-1",
            Hostname = "host1",
            RegisteredAt = DateTime.UtcNow,
            CurrentHealth = HealthStatus.Healthy
        });
        await db.SaveChangesAsync();

        // Use real CaService
        var controller = new AgentsController(db, _caService);

        var result = await controller.GetAgents(CancellationToken.None);

        var okResult = Assert.IsType<OkObjectResult>(result.Result);
        var agents = Assert.IsAssignableFrom<IReadOnlyList<AgentListDto>>(okResult.Value);
        Assert.Single(agents);
        Assert.Equal("agent-1", agents[0].Id);
        Assert.Equal("host1", agents[0].Hostname);
    }

    [Fact]
    public async Task GetAgents_WithHeartbeats_IncludesLatestSignals()
    {
        using var db = CreateInMemoryDb();
        db.Agents.Add(new Agent
        {
            Id = "agent-1",
            Hostname = "host1",
            RegisteredAt = DateTime.UtcNow,
            CurrentHealth = HealthStatus.Healthy
        });
        db.Heartbeats.AddRange(
            new Heartbeat
            {
                AgentId = "agent-1",
                Timestamp = DateTime.UtcNow.AddMinutes(-10),
                UptimeSeconds = 1000,
                LoadAverage1m = 1.5,
                MemoryUsedPct = 50,
                DiskUsedPct = 30,
                ProcessCount = 100,
                ReceivedAt = DateTime.UtcNow.AddMinutes(-10)
            },
            new Heartbeat
            {
                AgentId = "agent-1",
                Timestamp = DateTime.UtcNow.AddMinutes(-5),
                UptimeSeconds = 1300,
                LoadAverage1m = 2.0,
                MemoryUsedPct = 55,
                DiskUsedPct = 32,
                ProcessCount = 105,
                ReceivedAt = DateTime.UtcNow.AddMinutes(-5)
            }
        );
        await db.SaveChangesAsync();

        // Use real CaService
        var controller = new AgentsController(db, _caService);

        var result = await controller.GetAgents(CancellationToken.None);

        var okResult = Assert.IsType<OkObjectResult>(result.Result);
        var agents = Assert.IsAssignableFrom<IReadOnlyList<AgentListDto>>(okResult.Value);
        Assert.Single(agents);
        Assert.NotNull(agents[0].LatestSignals);
        Assert.Equal(2.0, agents[0].LatestSignals.LoadAverage1m);
        Assert.Equal(55, agents[0].LatestSignals.MemoryUsedPct);
    }

    [Fact]
    public async Task GetAgent_ExistingAgent_ReturnsAgentDetail()
    {
        using var db = CreateInMemoryDb();
        db.Agents.Add(new Agent
        {
            Id = "agent-1",
            Hostname = "host1",
            RegisteredAt = DateTime.UtcNow,
            CurrentHealth = HealthStatus.Healthy
        });
        await db.SaveChangesAsync();

        // Use real CaService
        var controller = new AgentsController(db, _caService);

        var result = await controller.GetAgent("agent-1", CancellationToken.None);

        var okResult = Assert.IsType<OkObjectResult>(result.Result);
        var agent = Assert.IsType<AgentDetailDto>(okResult.Value);
        Assert.Equal("agent-1", agent.Id);
        Assert.Equal("host1", agent.Hostname);
    }

    [Fact]
    public async Task GetAgent_NonexistentAgent_ReturnsNotFound()
    {
        using var db = CreateInMemoryDb();
        // Use real CaService
        var controller = new AgentsController(db, _caService);

        var result = await controller.GetAgent("nonexistent", CancellationToken.None);

        Assert.IsType<NotFoundResult>(result.Result);
    }

    [Fact]
    public async Task RegisterAgent_ValidRequest_ReturnsOk()
    {
        using var db = CreateInMemoryDb();
        var controller = new AgentsController(db, _caService);

        var csrPem = GenerateTestCsr();

        var request = new RegisterAgentRequest(
            "agent-123",
            "test-host",
            null,
            csrPem
        );

        var result = await controller.RegisterAgent(request, CancellationToken.None);

        var okResult = Assert.IsType<OkObjectResult>(result.Result);
        var response = Assert.IsType<RegisterAgentResponse>(okResult.Value);
        Assert.NotNull(response.CertPem);

        var agent = await db.Agents.FirstOrDefaultAsync(a => a.Id == "agent-123");
        Assert.NotNull(agent);
        Assert.Equal("test-host", agent.Hostname);
    }

    [Fact]
    public async Task RegisterAgent_MissingAgentId_ReturnsBadRequest()
    {
        using var db = CreateInMemoryDb();
        // Use real CaService
        var controller = new AgentsController(db, _caService);

        var request = new RegisterAgentRequest("", "test-host", null, "csr-pem");

        var result = await controller.RegisterAgent(request, CancellationToken.None);

        var badRequestResult = Assert.IsType<BadRequestObjectResult>(result.Result);
        Assert.Contains("agentId", badRequestResult.Value?.ToString());
    }

    [Fact]
    public async Task RegisterAgent_MissingHostname_ReturnsBadRequest()
    {
        using var db = CreateInMemoryDb();
        // Use real CaService
        var controller = new AgentsController(db, _caService);

        var request = new RegisterAgentRequest("agent-123", "", null, "csr-pem");

        var result = await controller.RegisterAgent(request, CancellationToken.None);

        var badRequestResult = Assert.IsType<BadRequestObjectResult>(result.Result);
        Assert.Contains("hostname", badRequestResult.Value?.ToString());
    }

    [Fact]
    public async Task RegisterAgent_InvalidCsr_ReturnsBadRequest()
    {
        using var db = CreateInMemoryDb();
        var controller = new AgentsController(db, _caService);

        var request = new RegisterAgentRequest("agent-123", "test-host", null, "invalid-csr");

        var result = await controller.RegisterAgent(request, CancellationToken.None);

        var badRequestResult = Assert.IsType<BadRequestObjectResult>(result.Result);
        Assert.Contains("csrPem invalid", badRequestResult.Value?.ToString());
    }

    [Fact]
    public async Task RevokeAgent_ExistingAgent_SetsRevokedAt()
    {
        using var db = CreateInMemoryDb();
        db.Agents.Add(new Agent
        {
            Id = "agent-1",
            Hostname = "host1",
            RegisteredAt = DateTime.UtcNow,
            CurrentHealth = HealthStatus.Healthy,
            RevokedAt = null
        });
        await db.SaveChangesAsync();

        // Use real CaService
        var controller = new AgentsController(db, _caService);

        var result = await controller.RevokeAgent("agent-1", CancellationToken.None);

        var okResult = Assert.IsType<OkObjectResult>(result.Result);
        var agent = Assert.IsType<Agent>(okResult.Value);
        Assert.NotNull(agent.RevokedAt);
    }

    [Fact]
    public async Task RevokeAgent_NonexistentAgent_ReturnsNotFound()
    {
        using var db = CreateInMemoryDb();
        // Use real CaService
        var controller = new AgentsController(db, _caService);

        var result = await controller.RevokeAgent("nonexistent", CancellationToken.None);

        Assert.IsType<NotFoundResult>(result.Result);
    }
}
