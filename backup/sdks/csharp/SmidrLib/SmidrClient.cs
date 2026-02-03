using Grpc.Core;
using Grpc.Net.Client;
using Smidr.V1;

namespace SmidrLib;

/// <summary>
/// Client for interacting with the Smidr gRPC daemon.
/// </summary>
public class SmidrClient : IDisposable
{
  private readonly GrpcChannel _channel;
  private readonly BuildService.BuildServiceClient _buildClient;
  private readonly LogService.LogServiceClient _logClient;
  private readonly ArtifactService.ArtifactServiceClient _artifactClient;

  /// <summary>
  /// Creates a new SmidrClient connected to the specified daemon address.
  /// </summary>
  /// <param name="address">The daemon address (e.g., "http://localhost:50051")</param>
  public SmidrClient(string address)
  {
    _channel = GrpcChannel.ForAddress(address);
    _buildClient = new BuildService.BuildServiceClient(_channel);
    _logClient = new LogService.LogServiceClient(_channel);
    _artifactClient = new ArtifactService.ArtifactServiceClient(_channel);
  }

  /// <summary>
  /// Gets the BuildService client for direct access to build operations.
  /// </summary>
  public BuildService.BuildServiceClient Builds => _buildClient;

  /// <summary>
  /// Gets the LogService client for direct access to log streaming operations.
  /// </summary>
  public LogService.LogServiceClient Logs => _logClient;

  /// <summary>
  /// Gets the ArtifactService client for direct access to artifact operations.
  /// </summary>
  public ArtifactService.ArtifactServiceClient Artifacts => _artifactClient;

  /// <summary>
  /// Starts a new build with the specified configuration.
  /// </summary>
  /// <param name="configPath">Path to the smidr.yaml configuration file</param>
  /// <param name="target">Build target (e.g., "core-image-minimal")</param>
  /// <param name="customer">Optional customer identifier</param>
  /// <param name="forceClean">Force a clean rebuild</param>
  /// <param name="forceImageRebuild">Force image rebuild only</param>
  /// <param name="cancellationToken">Cancellation token</param>
  /// <returns>Build status response with build ID and initial state</returns>
  public async Task<BuildStatusResponse> StartBuildAsync(
      string configPath,
      string target,
      string? customer = null,
      bool forceClean = false,
      bool forceImageRebuild = false,
      CancellationToken cancellationToken = default)
  {
    var request = new StartBuildRequest
    {
      Config = configPath,
      Target = target,
      Customer = customer ?? string.Empty,
      ForceClean = forceClean,
      ForceImageRebuild = forceImageRebuild
    };

    return await _buildClient.StartBuildAsync(request, cancellationToken: cancellationToken);
  }

  /// <summary>
  /// Gets the current status of a build.
  /// </summary>
  /// <param name="buildId">The build ID to query</param>
  /// <param name="cancellationToken">Cancellation token</param>
  /// <returns>Build status response with current state and metadata</returns>
  public async Task<BuildStatusResponse> GetBuildStatusAsync(
      string buildId,
      CancellationToken cancellationToken = default)
  {
    var request = new BuildStatusRequest
    {
      BuildIdentifier = new BuildIdentifier { BuildId = buildId }
    };

    return await _buildClient.GetBuildStatusAsync(request, cancellationToken: cancellationToken);
  }

  /// <summary>
  /// Lists all builds matching the specified filters.
  /// </summary>
  /// <param name="states">Optional list of build states to filter by</param>
  /// <param name="pageSize">Maximum number of builds to return (0 = all)</param>
  /// <param name="cancellationToken">Cancellation token</param>
  /// <returns>List of builds</returns>
  public async Task<ListBuildsResponse> ListBuildsAsync(
      IEnumerable<BuildState>? states = null,
      int pageSize = 0,
      CancellationToken cancellationToken = default)
  {
    var request = new ListBuildsRequest
    {
      PageSize = pageSize
    };

    if (states != null)
    {
      request.StateFilter.AddRange(states);
    }

    return await _buildClient.ListBuildsAsync(request, cancellationToken: cancellationToken);
  }

  /// <summary>
  /// Cancels a running build.
  /// </summary>
  /// <param name="buildId">The build ID to cancel</param>
  /// <param name="cancellationToken">Cancellation token</param>
  /// <returns>Cancellation response</returns>
  public async Task<CancelBuildResponse> CancelBuildAsync(
      string buildId,
      CancellationToken cancellationToken = default)
  {
    var request = new CancelBuildRequest
    {
      BuildIdentifier = new BuildIdentifier { BuildId = buildId }
    };

    return await _buildClient.CancelBuildAsync(request, cancellationToken: cancellationToken);
  }

  /// <summary>
  /// Lists artifacts for a completed build.
  /// </summary>
  /// <param name="buildId">The build ID to query</param>
  /// <param name="cancellationToken">Cancellation token</param>
  /// <returns>List of artifacts with metadata</returns>
  public async Task<ListArtifactsResponse> ListArtifactsAsync(
      string buildId,
      CancellationToken cancellationToken = default)
  {
    var request = new ListArtifactsRequest
    {
      BuildIdentifier = new BuildIdentifier { BuildId = buildId }
    };

    return await _artifactClient.ListArtifactsAsync(request, cancellationToken: cancellationToken);
  }

  /// <summary>
  /// Streams logs from a build in real-time.
  /// </summary>
  /// <param name="buildId">The build ID to stream logs from</param>
  /// <param name="follow">Whether to continue streaming new logs as they arrive</param>
  /// <param name="cancellationToken">Cancellation token</param>
  /// <returns>Async stream of log entries</returns>
  public async IAsyncEnumerable<LogEntry> StreamLogsAsync(
      string buildId,
      bool follow = false,
      [System.Runtime.CompilerServices.EnumeratorCancellation] CancellationToken cancellationToken = default)
  {
    var request = new StreamBuildLogsRequest
    {
      BuildIdentifier = new BuildIdentifier { BuildId = buildId },
      Follow = follow
    };

    using var call = _logClient.StreamBuildLogs(request, cancellationToken: cancellationToken);

    await foreach (var logEntry in call.ResponseStream.ReadAllAsync(cancellationToken))
    {
      yield return logEntry;
    }
  }

  /// <summary>
  /// Disposes the gRPC channel and releases resources.
  /// </summary>
  public void Dispose()
  {
    _channel.Dispose();
    GC.SuppressFinalize(this);
  }
}
