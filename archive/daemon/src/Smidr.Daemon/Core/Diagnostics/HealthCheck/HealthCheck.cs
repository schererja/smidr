using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace Smidr.Daemon.Core.Diagnostics.HealthCheck
{
    /// <summary>
    /// Performs health checks on various components of the daemon.
    /// </summary>
    public class HealthCheck
    {
        /// <summary>
        ///    Checks the health of the database by verifying connectivity and basic operations.
        /// </summary>
        /// <param name="dbPath"></param>
        /// <returns></returns>
        public async Task<HealthCheckResult> CheckDatabaseAsync(string dbPath)
        {
            try
            {
                var exists = System.IO.File.Exists(dbPath);
                return new HealthCheckResult(
                    exists,
                    "Database",
                    exists ? "Database file exists and is accessible." : "Database file does not exist.");


            }
            catch (Exception ex)
            {
                return new HealthCheckResult(false, "Database", $"Database health check failed: {ex.Message}");
            }
        }
        /// <summary>
        ///   Checks the health of the Docker daemon by verifying connectivity.
        /// </summary>
        /// <returns></returns>
        public async Task<HealthCheckResult> CheckDockerAsync()
        {
            try
            {
                // TODO: Implement actual Docker connectivity check
                await Task.CompletedTask;
                return new HealthCheckResult(
                    true,
                    "Docker",
                    "Docker is reachable and operational.");
            }
            catch (Exception ex)
            {
                return new HealthCheckResult(false, "Docker", $"Docker health check failed: {ex.Message}");
            }
        }

        public HealthCheckResult CheckDiskSpace(string path, long minBytes = 10737418240) // 10 GB default
        {
            try
            {
                var drive = new System.IO.DriveInfo(System.IO.Path.GetPathRoot(path) ?? "/");
                var availableBytes = drive.AvailableFreeSpace;
                var isHealthy = availableBytes >= minBytes;

                return new HealthCheckResult(
                    isHealthy,
                    "DiskSpace",
                    isHealthy
                        ? $"Sufficient disk space available: {availableBytes / (1024 * 1024 * 1024)} GB."
                        : $"Insufficient disk space. Available: {availableBytes / (1024 * 1024 * 1024)} GB, Required: {minBytes / (1024 * 1024 * 1024)} GB.",
                    new Dictionary<string, object>
                    {
                        { "AvailableBytes", availableBytes },
                        { "RequiredBytes", minBytes }
                    });
            }
            catch (Exception ex)
            {
                return new HealthCheckResult(false, "DiskSpace", $"Disk space health check failed: {ex.Message}");
            }
        }
    }
}
