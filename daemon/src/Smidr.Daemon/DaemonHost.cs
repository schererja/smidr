

using Docker.DotNet.Models;
using Grpc.Core;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Serilog;

namespace Smidr.Daemon
{
    public class DaemonHost : IHostedService
    {
        private readonly ILogger<DaemonHost> _logger;
        private readonly IConfiguration _configuration;
        private Server? _server;

        public DaemonHost(ILogger<DaemonHost> logger, IConfiguration configuration)
        {
            _logger = logger;
            _configuration = configuration;
        }
        public Task StartAsync(CancellationToken cancellationToken)
        {
            var host = _configuration["SMIDR_DAEMON_HOST"] ?? "localhost";
            var port = _configuration["SMIDR_DAEMON_PORT"] ?? "50051";

            _server = new Server
            {
                Ports =
              {
                new ServerPort(host, int.Parse(port), ServerCredentials.Insecure)
              }
            };
            _server.Start();
            _logger.LogInformation("gRPC server started on {Host}:{Port}", host, port);
            return Task.CompletedTask;
        }

        public async Task StopAsync(CancellationToken cancellationToken)
        {
            if (_server != null)
            {
                _logger.LogInformation("Shutting down gRPC server...");
                await _server.ShutdownAsync();
                _logger.LogInformation("gRPC server shut down.");
            }
        }
    }
}
