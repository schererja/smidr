using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace Smidr.Daemon.Core.Diagnostics.HealthCheck
{
    /// <summary>
    /// Represents the result of a health check for a specific component.
    /// </summary>
    public record HealthCheckResult
    (
        bool IsHealthy,
        string Component,
        string Message,
        Dictionary<string, object>? Metadata = null
    );
}
