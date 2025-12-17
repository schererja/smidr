namespace Smidr.Daemon.Core.Diagnostics.MetricsCollector
{
    public class DaemonMetrics
    {
        public long TotalBuildsStarted { get; set; }
        public long TotalBuildsCompleted { get; set; }
        public long TotalBuildsFailed { get; set; }
        public long UptimeSeconds { get; set; }
        public long CurrentMemoryMB { get; set; }
    }
}
