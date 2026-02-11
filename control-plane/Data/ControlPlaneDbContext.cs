using ControlPlane.Models;
using Microsoft.EntityFrameworkCore;

namespace ControlPlane.Data;

public sealed class ControlPlaneDbContext : DbContext
{
    public ControlPlaneDbContext(DbContextOptions<ControlPlaneDbContext> options)
        : base(options)
    {
    }

    public DbSet<Agent> Agents => Set<Agent>();
    public DbSet<Heartbeat> Heartbeats => Set<Heartbeat>();
    public DbSet<AgentBaseline> AgentBaselines => Set<AgentBaseline>();
    public DbSet<HealthState> HealthStates => Set<HealthState>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.Entity<Agent>(entity =>
        {
            entity.HasKey(agent => agent.Id);
            entity.Property(agent => agent.Hostname).IsRequired();
            entity.Property(agent => agent.RegisteredAt).IsRequired();
            entity.Property(agent => agent.CurrentHealth)
                .HasConversion<string>();
        });

        modelBuilder.Entity<Heartbeat>(entity =>
        {
            entity.HasKey(h => h.Id);
            entity.Property(h => h.AgentId).IsRequired();
            entity.HasIndex(h => h.AgentId);
            entity.HasIndex(h => h.Timestamp);
        });

        modelBuilder.Entity<AgentBaseline>(entity =>
        {
            entity.HasKey(b => b.Id);
            entity.Property(b => b.AgentId).IsRequired();
            entity.Property(b => b.MetricName).IsRequired();
            entity.HasIndex(b => new { b.AgentId, b.MetricName }).IsUnique();
        });

        modelBuilder.Entity<HealthState>(entity =>
        {
            entity.HasKey(h => h.Id);
            entity.Property(h => h.AgentId).IsRequired();
            entity.Property(h => h.Status)
                .HasConversion<string>();
            entity.HasIndex(h => h.AgentId);
            entity.HasIndex(h => h.ChangedAt);
        });
    }
}
