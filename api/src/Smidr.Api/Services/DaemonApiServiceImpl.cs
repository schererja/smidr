using Grpc.Core;
using Smidr.Proto.V1;

namespace Smidr.Api.Services;

public class DaemonApiServiceImpl : DaemonApiService.DaemonApiServiceBase
{

    private readonly ILogger<DaemonApiServiceImpl> _logger;
    private readonly IDaemonClient _daemonClient;

    /// <summary>
    /// Constructor for DaemonApiServiceImpl
    /// </summary>
    /// <param name="logger"></param>
    /// <param name="daemonClient"></param>
    public DaemonApiServiceImpl(ILogger<DaemonApiServiceImpl> logger, IDaemonClient daemonClient)
    {
        _logger = logger;
        _daemonClient = daemonClient;
    }
    /// <summary>
    /// Handles SubmitTask gRPC requests. Forwards the request to the DaemonClient and returns the response.
    /// References: docs/architecture/TaskLifecycle.md
    /// </summary>
    /// <param name="request"></param>
    /// <param name="context"></param>
    /// <returns></returns>
    public override async Task<SubmitTaskResponse> SubmitTask(SubmitTaskRequest request, ServerCallContext context)
    {
        _logger.LogInformation("Received SubmitTask request for TaskId: {TaskId}", request.TaskId);

        var response = await _daemonClient.SubmitTaskAsync(request);

        _logger.LogInformation("TaskId: {TaskId} submitted successfully with Status: {Status}", request.TaskId, response.Status);

        return response;
    }

    public override async Task<ListTasksResponse> ListTasks(ListTasksRequest request, ServerCallContext context)
    {
        _logger.LogInformation("Received ListTasks request");

        var response = await _daemonClient.ListTasksAsync(request);

        _logger.LogInformation("ListTasks request completed with {TaskCount} tasks", response.Tasks.Count);

        return response;
    }



}
