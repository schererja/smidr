using SmidrLib;
using Smidr.V1;

var builder = WebApplication.CreateBuilder(args);

// Add services to the container
builder.Services.AddOpenApi();

// Configure gRPC client for Smidr daemon
var daemonUrl = builder.Configuration["Smidr:DaemonUrl"] ?? "http://localhost:50051";
builder.Services.AddSingleton(new SmidrClient(daemonUrl));

var app = builder.Build();

// Configure the HTTP request pipeline
if (app.Environment.IsDevelopment())
{
  app.MapOpenApi();
}

app.UseHttpsRedirection();

// Build Management Endpoints

app.MapPost("/api/builds", async (StartBuildDto dto, SmidrClient client) =>
{
  var response = await client.StartBuildAsync(
      configPath: dto.ConfigPath,
      target: dto.Target,
      customer: dto.Customer,
      forceClean: dto.ForceClean ?? false,
      forceImageRebuild: dto.ForceImageRebuild ?? false
  );

  return Results.Ok(new
  {
    buildId = response.BuildIdentifier.BuildId,
    message = response.Message
  });
})
.WithName("StartBuild")
.WithSummary("Start a new Yocto build")
.WithOpenApi();

app.MapGet("/api/builds/{buildId}", async (string buildId, SmidrClient client) =>
{
  try
  {
    var status = await client.GetBuildStatusAsync(buildId);
    return Results.Ok(new
    {
      buildId = status.BuildIdentifier.BuildId,
      state = status.State.ToString(),
      startTime = DateTimeOffset.FromUnixTimeSeconds(status.Timestamps.StartTimeUnixSeconds),
      endTime = status.Timestamps.EndTimeUnixSeconds > 0
            ? DateTimeOffset.FromUnixTimeSeconds(status.Timestamps.EndTimeUnixSeconds)
            : (DateTimeOffset?)null,
      errorMessage = status.ErrorMessage,
      exitCode = status.ExitCode
    });
  }
  catch (Grpc.Core.RpcException ex) when (ex.StatusCode == Grpc.Core.StatusCode.NotFound)
  {
    return Results.NotFound(new { error = "Build not found" });
  }
})
.WithName("GetBuildStatus")
.WithSummary("Get the status of a specific build")
.WithOpenApi();

app.MapGet("/api/builds", async (SmidrClient client, int? pageSize = null, string? state = null) =>
{
  BuildState[]? states = null;
  if (!string.IsNullOrEmpty(state) && Enum.TryParse<BuildState>(state, out var parsedState))
  {
    states = new[] { parsedState };
  }

  var response = await client.ListBuildsAsync(states: states, pageSize: pageSize);

  return Results.Ok(new
  {
    builds = response.Builds.Select(b => new
    {
      buildId = b.BuildIdentifier.BuildId,
      state = b.State.ToString(),
      startTime = DateTimeOffset.FromUnixTimeSeconds(b.Timestamps.StartTimeUnixSeconds),
      target = b.Target
    })
  });
})
.WithName("ListBuilds")
.WithSummary("List all builds with optional filtering")
.WithOpenApi();

app.MapDelete("/api/builds/{buildId}", async (string buildId, SmidrClient client) =>
{
  try
  {
    var response = await client.CancelBuildAsync(buildId);
    return Results.Ok(new { success = response.Success, message = response.Message });
  }
  catch (Grpc.Core.RpcException ex) when (ex.StatusCode == Grpc.Core.StatusCode.NotFound)
  {
    return Results.NotFound(new { error = "Build not found" });
  }
})
.WithName("CancelBuild")
.WithSummary("Cancel a running build")
.WithOpenApi();

// Artifact Endpoints

app.MapGet("/api/builds/{buildId}/artifacts", async (string buildId, SmidrClient client) =>
{
  try
  {
    var response = await client.ListArtifactsAsync(buildId);
    return Results.Ok(new
    {
      buildId,
      artifacts = response.Artifacts.Select(a => new
      {
        name = a.Name,
        path = a.Path,
        sizeBytes = a.SizeBytes,
        downloadUrl = a.DownloadUrl,
        checksum = a.Checksum
      })
    });
  }
  catch (Grpc.Core.RpcException ex) when (ex.StatusCode == Grpc.Core.StatusCode.NotFound)
  {
    return Results.NotFound(new { error = "Build not found" });
  }
})
.WithName("ListArtifacts")
.WithSummary("List artifacts for a completed build")
.WithOpenApi();

// Log Streaming Endpoint (Server-Sent Events)

app.MapGet("/api/builds/{buildId}/logs", async (string buildId, bool follow, SmidrClient client, HttpContext context) =>
{
  context.Response.Headers.Append("Content-Type", "text/event-stream");
  context.Response.Headers.Append("Cache-Control", "no-cache");
  context.Response.Headers.Append("Connection", "keep-alive");

  try
  {
    await foreach (var logEntry in client.StreamLogsAsync(buildId, follow, context.RequestAborted))
    {
      var data = $"data: {System.Text.Json.JsonSerializer.Serialize(new { stream = logEntry.Stream, message = logEntry.Message, timestamp = logEntry.TimestampUnixSeconds })}\n\n";
      await context.Response.WriteAsync(data, context.RequestAborted);
      await context.Response.Body.FlushAsync(context.RequestAborted);
    }
  }
  catch (OperationCanceledException)
  {
    // Client disconnected
  }
  catch (Grpc.Core.RpcException ex) when (ex.StatusCode == Grpc.Core.StatusCode.NotFound)
  {
    await context.Response.WriteAsync("event: error\ndata: Build not found\n\n");
  }
})
.WithName("StreamLogs")
.WithSummary("Stream build logs in real-time using Server-Sent Events")
.WithOpenApi()
.ExcludeFromDescription(); // SSE doesn't work well with OpenAPI

app.Run();

// DTOs
record StartBuildDto(
    string ConfigPath,
    string Target,
    string? Customer = null,
    bool? ForceClean = null,
    bool? ForceImageRebuild = null
);
