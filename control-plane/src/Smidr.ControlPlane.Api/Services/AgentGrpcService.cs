using Grpc.Core;
using Microsoft.EntityFrameworkCore;
using Smidr.Agent;
using Smidr.ControlPlane.Application.Common;
using Smidr.ControlPlane.Infrastructure.Persistence;
using Smidr.Common;
using AgentState = Smidr.ControlPlane.Domain.Entities.AgentState;

namespace Smidr.ControlPlane.Api.Services;

public class AgentGrpcService : AgentService.AgentServiceBase
{
    private readonly SmidrDbContext _dbContext;
    private readonly ITenantContext _tenantContext;
    private readonly ILogger<AgentGrpcService> _logger;

    public AgentGrpcService(
        SmidrDbContext dbContext,
        ITenantContext tenantContext,
        ILogger<AgentGrpcService> logger)
    {
        _dbContext = dbContext;
        _tenantContext = tenantContext;
        _logger = logger;
    }

    public override async Task<RegisterAgentResponse> RegisterAgent(
        RegisterAgentRequest request,
        ServerCallContext context)
    {
        try
        {
            if (!_tenantContext.IsSet)
            {
                throw new RpcException(new Status(StatusCode.Unauthenticated, "Tenant not identified"));
            }

            var agentId = Guid.NewGuid();
            var agent = new Domain.Entities.Agent
            {
                Id = agentId,
                TenantId = _tenantContext.TenantId,
                Name = request.Name,
                Hostname = request.Hostname,
                State = AgentState.Registering,
                RegisteredAt = DateTime.UtcNow,
                LastHeartbeatAt = DateTime.UtcNow,
                Platform = request.Capabilities?.Platform.ToString() ?? "Unknown",
                Architecture = request.Capabilities?.Architecture.ToString() ?? "Unknown",
                Labels = request.Capabilities?.Labels.ToDictionary(k => k.Key, v => v.Value) ?? new(),
                Features = request.Capabilities?.Features.ToList() ?? new(),
                CpuMillicores = 0,
                MemoryBytes = 0,
                DiskBytes = 0
            };

            _dbContext.Agents.Add(agent);
            await _dbContext.SaveChangesAsync();

            agent.State = AgentState.Idle;
            await _dbContext.SaveChangesAsync();

            _logger.LogInformation("Agent {AgentId} registered for tenant {TenantId}", agentId, _tenantContext.TenantId);

            return new RegisterAgentResponse
            {
                AgentId = new Smidr.Common.AgentID { Value = agentId.ToString() },
                Success = true,
                Message = "Agent registered successfully"
            };
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error registering agent");
            throw new RpcException(new Status(StatusCode.Internal, ex.Message));
        }
    }

    public override async Task<HeartbeatResponse> Heartbeat(
        HeartbeatRequest request,
        ServerCallContext context)
    {
        try
        {
            if (!Guid.TryParse(request.AgentId.Value, out var agentId))
            {
                throw new RpcException(new Status(StatusCode.InvalidArgument, "Invalid agent ID"));
            }

            var agent = await _dbContext.Agents
                .FirstOrDefaultAsync(a => a.Id == agentId && a.TenantId == _tenantContext.TenantId);

            if (agent == null)
            {
                throw new RpcException(new Status(StatusCode.NotFound, "Agent not found"));
            }

            agent.LastHeartbeatAt = DateTime.UtcNow;
            agent.State = request.State switch
            {
                Smidr.Common.AgentState.Busy => AgentState.Busy,
                Smidr.Common.AgentState.Idle => AgentState.Idle,
                Smidr.Common.AgentState.Registering => AgentState.Registering,
                Smidr.Common.AgentState.Offline => AgentState.Offline,
                Smidr.Common.AgentState.Decommissioned => AgentState.Decommissioned,
                _ => agent.State
            };
            await _dbContext.SaveChangesAsync();

            return new HeartbeatResponse
            {
                Alive = true
            };
        }
        catch (RpcException)
        {
            throw;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error processing heartbeat");
            throw new RpcException(new Status(StatusCode.Internal, ex.Message));
        }
    }

    public override async Task<GetJobAssignmentResponse> GetJobAssignment(
        GetJobAssignmentRequest request,
        ServerCallContext context)
    {
        try
        {
            if (!Guid.TryParse(request.AgentId.Value, out var agentId))
            {
                throw new RpcException(new Status(StatusCode.InvalidArgument, "Invalid agent ID"));
            }

            var agent = await _dbContext.Agents
                .FirstOrDefaultAsync(a => a.Id == agentId && a.TenantId == _tenantContext.TenantId);

            if (agent == null)
            {
                throw new RpcException(new Status(StatusCode.NotFound, "Agent not found"));
            }

            // Find queued job matching agent capabilities
            var job = await _dbContext.Jobs
                .Where(j => j.TenantId == _tenantContext.TenantId &&
                           j.State == Domain.Entities.JobState.Queued &&
                           j.AssignedAgentId == null)
                .OrderBy(j => j.SubmittedAt)
                .FirstOrDefaultAsync();

            var response = new GetJobAssignmentResponse();
            if (job == null)
            {
                return response;
            }

            // Assign job to agent
            job.AssignedAgentId = agentId;
            job.State = Domain.Entities.JobState.Dispatched;
            agent.State = AgentState.Busy;
            await _dbContext.SaveChangesAsync();

            _logger.LogInformation("Job {JobId} dispatched to agent {AgentId}", job.Id, agentId);

            response.Jobs.Add(new Smidr.Agent.JobAssignment
            {
                JobId = new Smidr.Common.JobID { Value = job.Id.ToString() },
                JobType = job.JobType,
                Deadline = Google.Protobuf.WellKnownTypes.Timestamp.FromDateTime(DateTime.UtcNow.AddHours(1))
            });
            return response;
        }
        catch (RpcException)
        {
            throw;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error fetching job");
            throw new RpcException(new Status(StatusCode.Internal, ex.Message));
        }
    }

    public override async Task<ReportJobCompletionResponse> ReportJobCompletion(
        ReportJobCompletionRequest request,
        ServerCallContext context)
    {
        try
        {
            if (!Guid.TryParse(request.JobId.Value, out var jobId))
            {
                throw new RpcException(new Status(StatusCode.InvalidArgument, "Invalid job ID"));
            }

            var job = await _dbContext.Jobs
                .Include(j => j.AssignedAgent)
                .FirstOrDefaultAsync(j => j.Id == jobId && j.TenantId == _tenantContext.TenantId);

            if (job == null)
            {
                throw new RpcException(new Status(StatusCode.NotFound, "Job not found"));
            }

            job.State = request.FinalState switch
            {
                Smidr.Common.JobState.Succeeded => Domain.Entities.JobState.Succeeded,
                Smidr.Common.JobState.Failed => Domain.Entities.JobState.Failed,
                Smidr.Common.JobState.Cancelled => Domain.Entities.JobState.Cancelled,
                Smidr.Common.JobState.TimedOut => Domain.Entities.JobState.TimedOut,
                Smidr.Common.JobState.Running => Domain.Entities.JobState.Running,
                Smidr.Common.JobState.Dispatched => Domain.Entities.JobState.Dispatched,
                Smidr.Common.JobState.Queued => Domain.Entities.JobState.Queued,
                Smidr.Common.JobState.Submitted => Domain.Entities.JobState.Submitted,
                _ => job.State
            };
            job.CompletedAt = DateTime.UtcNow;
            job.ExitCode = request.ExitCode;
            job.ErrorMessage = request.ErrorMessage;

            if (job.AssignedAgent != null)
            {
                job.AssignedAgent.State = AgentState.Idle;
            }

            await _dbContext.SaveChangesAsync();

            _logger.LogInformation("Job {JobId} completed with status {Status}", jobId, job.State);

            return new ReportJobCompletionResponse
            {
                Acknowledged = true,
                Message = "Result received"
            };
        }
        catch (RpcException)
        {
            throw;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error reporting job result");
            throw new RpcException(new Status(StatusCode.Internal, ex.Message));
        }
    }
}
