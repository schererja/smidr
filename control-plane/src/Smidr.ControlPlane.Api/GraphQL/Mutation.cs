using HotChocolate;
using Smidr.ControlPlane.Application.Common;
using Smidr.ControlPlane.Domain.Entities;
using Smidr.ControlPlane.Infrastructure.Persistence;

namespace Smidr.ControlPlane.Api.GraphQL;

public class Mutation
{
    [UseDbContext(typeof(SmidrDbContext))]
    public async Task<Smidr.ControlPlane.Domain.Entities.Tenant> CreateTenant(
        string name,
        string slug,
        [Service(ServiceKind.Pooled)] SmidrDbContext dbContext)
    {
        var tenant = new Smidr.ControlPlane.Domain.Entities.Tenant
        {
            Id = Guid.NewGuid(),
            Name = name,
            Slug = slug,
            Status = Smidr.ControlPlane.Domain.Entities.TenantStatus.Active,
            CreatedAt = DateTime.UtcNow
        };

        dbContext.Tenants.Add(tenant);
        await dbContext.SaveChangesAsync();

        return tenant;
    }

    [UseDbContext(typeof(SmidrDbContext))]
    public async Task<Smidr.ControlPlane.Domain.Entities.Project> CreateProject(
        string name,
        string description,
        [Service(ServiceKind.Pooled)] SmidrDbContext dbContext,
        [Service] ITenantContext tenantContext)
    {
        if (!tenantContext.IsSet)
        {
            throw new GraphQLException("Tenant not identified");
        }

        var project = new Smidr.ControlPlane.Domain.Entities.Project
        {
            Id = Guid.NewGuid(),
            TenantId = tenantContext.TenantId,
            Name = name,
            Description = description,
            CreatedAt = DateTime.UtcNow
        };

        dbContext.Projects.Add(project);
        await dbContext.SaveChangesAsync();

        return project;
    }

    [UseDbContext(typeof(SmidrDbContext))]
    public async Task<Smidr.ControlPlane.Domain.Entities.Job> SubmitJob(
        Guid projectId,
        string name,
        string jobType,
        string description,
        [Service(ServiceKind.Pooled)] SmidrDbContext dbContext,
        [Service] ITenantContext tenantContext)
    {
        if (!tenantContext.IsSet)
        {
            throw new GraphQLException("Tenant not identified");
        }

        var job = new Smidr.ControlPlane.Domain.Entities.Job
        {
            Id = Guid.NewGuid(),
            TenantId = tenantContext.TenantId,
            ProjectId = projectId,
            CreatedByUserId = Guid.Empty, // TODO: Get from user context
            Name = name,
            JobType = jobType,
            Description = description,
            State = Smidr.ControlPlane.Domain.Entities.JobState.Queued,
            SubmittedAt = DateTime.UtcNow,
            MaxRetries = 3
        };

        dbContext.Jobs.Add(job);
        await dbContext.SaveChangesAsync();

        return job;
    }

    [UseDbContext(typeof(SmidrDbContext))]
    public async Task<Smidr.ControlPlane.Domain.Entities.Job?> CancelJob(
        Guid id,
        [Service(ServiceKind.Pooled)] SmidrDbContext dbContext,
        [Service] ITenantContext tenantContext)
    {
        if (!tenantContext.IsSet)
        {
            throw new GraphQLException("Tenant not identified");
        }

        var job = await dbContext.Jobs.FindAsync(id);
        if (job == null || job.TenantId != tenantContext.TenantId)
        {
            return null;
        }

        if (job.State == Smidr.ControlPlane.Domain.Entities.JobState.Queued || job.State == Smidr.ControlPlane.Domain.Entities.JobState.Dispatched || job.State == Smidr.ControlPlane.Domain.Entities.JobState.Running)
        {
            job.State = Smidr.ControlPlane.Domain.Entities.JobState.Cancelled;
            job.CompletedAt = DateTime.UtcNow;
            await dbContext.SaveChangesAsync();
        }

        return job;
    }
}
