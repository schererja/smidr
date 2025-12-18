using System;
using System.Threading;
using System.Threading.Tasks;
using Grpc.Net.Client;
using Smidr.Proto.V1;

namespace Smidr.Daemon.Services.Grpc;

public class DaemonApiClient
{
    private readonly DaemonApiService.DaemonApiServiceClient _client;

    public DaemonApiClient(string baseAddress)
    {
        var channel = GrpcChannel.ForAddress(baseAddress);
        _client = new DaemonApiService.DaemonApiServiceClient(channel);
    }

    public async Task<string> RegisterAsync(string hostname, string[] providers)
    {
        var request = new RegisterDaemonRequest
        {
            Hostname = hostname,
            Version = "0.1.0"
        };
        request.SupportedProviders.AddRange(providers);
        var response = await _client.RegisterDaemonAsync(request);
        return response.AssignedId;
    }

    public async Task SendHeartbeatAsync(string daemonId)
    {
        var hb = new HeartbeatRequest
        {
            DaemonId = daemonId,
            At = Google.Protobuf.WellKnownTypes.Timestamp.FromDateTime(DateTime.UtcNow),
            CpuUtilization = 0,
            MemoryUtilization = 0
        };
        await _client.HeartbeatAsync(hb);
    }

    public async Task ReceiveTasksAsync(string daemonId, CancellationToken cancellationToken)
    {
        var sub = new ReceiveTasksRequest { DaemonId = daemonId };
        using var call = _client.ReceiveTasks(sub);
        while (await call.ResponseStream.MoveNext(cancellationToken))
        {
            var item = call.ResponseStream.Current;
            var instruction = item.Instruction;
            // TODO: route to scheduler/provider
        }
    }

    public Task AcknowledgeTaskAsync(string taskId, bool accepted, string message)
    {
        var req = new AcknowledgeTaskRequest { TaskId = taskId, Accepted = accepted, Message = message };
        return _client.AcknowledgeTaskAsync(req).ResponseAsync;
    }

    public Task ReportTaskUpdateAsync(string taskId, TaskState state, string reason)
    {
        var req = new ReportTaskUpdateRequest
        {
            TaskId = taskId,
            State = state,
            Reason = reason,
            At = Google.Protobuf.WellKnownTypes.Timestamp.FromDateTime(DateTime.UtcNow)
        };
        return _client.ReportTaskUpdateAsync(req).ResponseAsync;
    }

    public Task ReportMetricsAsync(string daemonId)
    {
        var req = new ReportMetricsRequest
        {
            DaemonId = daemonId,
            At = Google.Protobuf.WellKnownTypes.Timestamp.FromDateTime(DateTime.UtcNow)
        };
        return _client.ReportMetricsAsync(req).ResponseAsync;
    }
}
