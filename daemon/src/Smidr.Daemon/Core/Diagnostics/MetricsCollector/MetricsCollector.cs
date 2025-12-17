using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;
using System.Threading;

namespace Smidr.Daemon.Core.Diagnostics.MetricsCollector
{
    public class MetricsCollector
    {
        private long _totalBuildsStarted;
        private long _totalBuildsCompleted;
        private long _totalBuildsFailed;
        private readonly Stopwatch _uptimeStopwatch;

        public MetricsCollector()
        {
            _uptimeStopwatch = Stopwatch.StartNew();
        }

        public void RecordBuildStarted()
        {
            Interlocked.Increment(ref _totalBuildsStarted);
        }
        public void RecordBuildCompleted()
        {
            Interlocked.Increment(ref _totalBuildsCompleted);
        }
        public void RecordBuildFailed()
        {
            Interlocked.Increment(ref _totalBuildsFailed);
        }
        public DaemonMetrics GetMetrics()
        {
            return new DaemonMetrics
            {
                TotalBuildsStarted = _totalBuildsStarted,
                TotalBuildsCompleted = _totalBuildsCompleted,
                TotalBuildsFailed = _totalBuildsFailed,
                UptimeSeconds = (long)_uptimeStopwatch.Elapsed.TotalSeconds,
                CurrentMemoryMB = Process.GetCurrentProcess().WorkingSet64 / 1024 / 1024
            };
        }
    }
}
