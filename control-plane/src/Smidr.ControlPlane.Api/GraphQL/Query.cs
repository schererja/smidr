using HotChocolate;
using Microsoft.EntityFrameworkCore;
using Smidr.ControlPlane.Application.Common;
using Smidr.ControlPlane.Domain.Entities;
using Smidr.ControlPlane.Infrastructure.Persistence;

namespace Smidr.ControlPlane.Api.GraphQL;

public class Query
{
    [UseDbContext(typeof(SmidrDbContext))]
    [UseFiltering]
    [UseSorting]
    public IQueryable<Smidr.ControlPlane.Domain.Entities.Tenant> GetTenants([Service(ServiceKind.Pooled)] SmidrDbContext dbContext)
    {
        return dbContext.Tenants;
    }

    [UseDbContext(typeof(SmidrDbContext))]
    public async Task<Smidr.ControlPlane.Domain.Entities.Tenant?> GetTenant(
        Guid id,
        [Service(ServiceKind.Pooled)] SmidrDbContext dbContext)
    {
        return await dbContext.Tenants.FindAsync(id);
    }

    [UseDbContext(typeof(SmidrDbContext))]
    [UseFiltering]
    [UseSorting]
    public IQueryable<Smidr.ControlPlane.Domain.Entities.Agent> GetAgents(
        [Service(ServiceKind.Pooled)] SmidrDbContext dbContext,
        [Service] ITenantContext tenantContext)
    {
        if (!tenantContext.IsSet)
        {
            throw new GraphQLException("Tenant not identified");
        }
        return dbContext.Agents.Where(a => a.TenantId == tenantContext.TenantId);
    }

    [UseDbContext(typeof(SmidrDbContext))]
    [UseFiltering]
    [UseSorting]
    public IQueryable<Smidr.ControlPlane.Domain.Entities.Job> GetJobs(
        [Service(ServiceKind.Pooled)] SmidrDbContext dbContext,
        [Service] ITenantContext tenantContext)
    {
        if (!tenantContext.IsSet)
        {
            throw new GraphQLException("Tenant not identified");
        }
        return dbContext.Jobs
            .Where(j => j.TenantId == tenantContext.TenantId)
            .Include(j => j.AssignedAgent)
            .Include(j => j.Project);
    }

    [UseDbContext(typeof(SmidrDbContext))]
    public async Task<Smidr.ControlPlane.Domain.Entities.Job?> GetJob(
        Guid id,
        [Service(ServiceKind.Pooled)] SmidrDbContext dbContext,
        [Service] ITenantContext tenantContext)
    {
        if (!tenantContext.IsSet)
        {
            throw new GraphQLException("Tenant not identified");
        }
        return await dbContext.Jobs
            .Include(j => j.AssignedAgent)
            .Include(j => j.Project)
            .Include(j => j.Artifacts)
            .FirstOrDefaultAsync(j => j.Id == id && j.TenantId == tenantContext.TenantId);
    }

    [UseDbContext(typeof(SmidrDbContext))]
    [UseFiltering]
    [UseSorting]
    public IQueryable<Smidr.ControlPlane.Domain.Entities.Project> GetProjects(
        [Service(ServiceKind.Pooled)] SmidrDbContext dbContext,
        [Service] ITenantContext tenantContext)
    {
        if (!tenantContext.IsSet)
        {
            throw new GraphQLException("Tenant not identified");
        }
        return dbContext.Projects.Where(p => p.TenantId == tenantContext.TenantId);
    }
}
