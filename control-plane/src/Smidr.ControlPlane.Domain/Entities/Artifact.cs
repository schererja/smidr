namespace Smidr.ControlPlane.Domain.Entities;

public class Artifact
{
    public Guid Id { get; set; }
    public Guid TenantId { get; set; }
    public Guid ProjectId { get; set; }
    public Guid JobId { get; set; }
    public Guid AgentId { get; set; }

    public string Name { get; set; } = string.Empty;
    public string ArtifactType { get; set; } = string.Empty;
    public string Version { get; set; } = string.Empty;
    public string StoragePath { get; set; } = string.Empty;

    public long SizeBytes { get; set; }
    public string Checksum { get; set; } = string.Empty;
    public string ChecksumAlgorithm { get; set; } = "SHA256";

    public ArtifactStage Stage { get; set; }
    public DateTime CreatedAt { get; set; }
    public DateTime? RetentionUntil { get; set; }

    // Relationships
    public Job Job { get; set; } = null!;
}

public enum ArtifactStage
{
    Development,
    QA,
    Staging,
    Production,
    Archived
}
