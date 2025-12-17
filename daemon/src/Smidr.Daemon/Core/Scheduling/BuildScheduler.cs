using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;
namespace Smidr.Daemon.Core.Scheduling
{
    /// <summary>
    /// Schedules builds based on priority, resource availability, and customer constraints.
    /// </summary>
    public class BuildScheduler
    {
        private readonly BuildQueue _queue;
        private readonly ILogger<BuildScheduler> _logger;
        private readonly CancellationTokenSource _cancellationTokenSource;
        public BuildScheduler(BuildQueue queue, ILogger<BuildScheduler> logger)
        {
            _queue = queue ?? throw new ArgumentNullException(nameof(queue));
            _logger = logger ?? throw new ArgumentNullException(nameof(logger));
            _cancellationTokenSource = new CancellationTokenSource();
        }
        public async Task<string> ScheduleBuildAsync(string customerId, string target, string configPath, CancellationToken cancellationToken)
        {
            var buildId = Guid.NewGuid().ToString();
            var queuedBuild = new QueuedBuild(buildId, customerId, target, configPath, DateTime.UtcNow);

            var enqueueSuccess = await _queue.EnqueueAsync(queuedBuild, cancellationToken);
            if (!enqueueSuccess)
            {
                _logger.LogWarning("Failed to enqueue build {BuildId} for customer {CustomerId}", buildId, customerId);
                throw new OperationCanceledException("Build scheduling was cancelled.");
            }

            _logger.LogInformation("Build {BuildId} for customer {CustomerId} scheduled successfully", buildId, customerId);

            // Start build execution in the background
            _ = Task.Run(async () => await ExecuteBuildAsync(queuedBuild, _cancellationTokenSource.Token), _cancellationTokenSource.Token);
            return buildId;
        }

        private async Task ExecuteBuildAsync(QueuedBuild build, CancellationToken cancellationToken)
        {
            try
            {
                _logger.LogInformation($"Starting execution of build {build.BuildId} for customer {build.CustomerId}");
                // TODO: Implement actual build execution logic here
                await Task.Delay(TimeSpan.FromMinutes(5), cancellationToken); // Simulate build time
                _logger.LogInformation($"Build {build.BuildId} for customer {build.CustomerId} completed successfully");
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Error during execution of build {BuildId} for customer {CustomerId}", build.BuildId, build.CustomerId);
            }
            finally
            {
                _queue.ReleaseBuild(build.CustomerId);
                _logger.LogInformation("Released resources for build {BuildId} of customer {CustomerIds}", build.BuildId, build.CustomerId);
            }

        }

        public void Shutdown()
        {
            _cancellationTokenSource.Cancel();
        }
    }
}
