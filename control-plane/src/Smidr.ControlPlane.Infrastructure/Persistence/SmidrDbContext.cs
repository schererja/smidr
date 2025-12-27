using Microsoft.EntityFrameworkCore;
using Smidr.ControlPlane.Domain.Entities;

namespace Smidr.ControlPlane.Infrastructure.Persistence;

public class SmidrDbContext : DbContext
{
    public SmidrDbContext(DbContextOptions<SmidrDbContext> options) : base(options)
    {
    }

    public DbSet<Tenant> Tenants => Set<Tenant>();
    public DbSet<Agent> Agents => Set<Agent>();
    public DbSet<Job> Jobs => Set<Job>();
    public DbSet<Project> Projects => Set<Project>();
    public DbSet<Artifact> Artifacts => Set<Artifact>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        base.OnModelCreating(modelBuilder);

        // Tenant
        modelBuilder.Entity<Tenant>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.Slug).IsUnique();
            entity.Property(e => e.Name).IsRequired().HasMaxLength(200);
            entity.Property(e => e.Slug).IsRequired().HasMaxLength(100);
        });

        // Agent
        modelBuilder.Entity<Agent>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => new { e.TenantId, e.State });
            entity.Property(e => e.Name).IsRequired().HasMaxLength(200);
            entity.Property(e => e.Labels).HasConversion(
                v => System.Text.Json.JsonSerializer.Serialize(v, (System.Text.Json.JsonSerializerOptions?)null),
                v => System.Text.Json.JsonSerializer.Deserialize<Dictionary<string, string>>(v, (System.Text.Json.JsonSerializerOptions?)null) ?? new());
            entity.Property(e => e.Features).HasConversion(
                v => System.Text.Json.JsonSerializer.Serialize(v, (System.Text.Json.JsonSerializerOptions?)null),
                v => System.Text.Json.JsonSerializer.Deserialize<List<string>>(v, (System.Text.Json.JsonSerializerOptions?)null) ?? new());

            entity.HasOne(e => e.Tenant)
                .WithMany(t => t.Agents)
                .HasForeignKey(e => e.TenantId)
                .OnDelete(DeleteBehavior.Cascade);
        });

        // Job
        modelBuilder.Entity<Job>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => new { e.TenantId, e.State });
            entity.HasIndex(e => e.ProjectId);
            entity.Property(e => e.Name).IsRequired().HasMaxLength(200);
            entity.Property(e => e.Parameters).HasConversion(
                v => System.Text.Json.JsonSerializer.Serialize(v, (System.Text.Json.JsonSerializerOptions?)null),
                v => System.Text.Json.JsonSerializer.Deserialize<Dictionary<string, string>>(v, (System.Text.Json.JsonSerializerOptions?)null) ?? new());
            entity.Property(e => e.RequiredCapabilities).HasConversion(
                v => System.Text.Json.JsonSerializer.Serialize(v, (System.Text.Json.JsonSerializerOptions?)null),
                v => System.Text.Json.JsonSerializer.Deserialize<List<string>>(v, (System.Text.Json.JsonSerializerOptions?)null) ?? new());

            entity.HasOne(e => e.Tenant)
                .WithMany(t => t.Jobs)
                .HasForeignKey(e => e.TenantId)
                .OnDelete(DeleteBehavior.Cascade);

            entity.HasOne(e => e.Project)
                .WithMany(p => p.Jobs)
                .HasForeignKey(e => e.ProjectId)
                .OnDelete(DeleteBehavior.Restrict);

            entity.HasOne(e => e.AssignedAgent)
                .WithMany(a => a.AssignedJobs)
                .HasForeignKey(e => e.AssignedAgentId)
                .OnDelete(DeleteBehavior.SetNull);
        });

        // Project
        modelBuilder.Entity<Project>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.TenantId);
            entity.Property(e => e.Name).IsRequired().HasMaxLength(200);

            entity.HasOne(e => e.Tenant)
                .WithMany(t => t.Projects)
                .HasForeignKey(e => e.TenantId)
                .OnDelete(DeleteBehavior.Cascade);
        });

        // Artifact
        modelBuilder.Entity<Artifact>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => new { e.TenantId, e.JobId });
            entity.Property(e => e.Name).IsRequired().HasMaxLength(200);

            entity.HasOne(e => e.Job)
                .WithMany(j => j.Artifacts)
                .HasForeignKey(e => e.JobId)
                .OnDelete(DeleteBehavior.Cascade);
        });
    }
}
