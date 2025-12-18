
using System.Collections.Concurrent;

namespace Smidr.Daemon.Core.Scheduling
{
    public class BuildQueue
    {
        private readonly SemaphoreSlim _semaphore;
        private readonly ConcurrentDictionary<string, SemaphoreSlim> _customerSemaphores;
        private readonly ConcurrentQueue<QueuedBuild> _queue;
        public BuildQueue(int maxConcurrentBuilds)
        {
            _semaphore = new SemaphoreSlim(maxConcurrentBuilds, maxConcurrentBuilds);
            _customerSemaphores = new ConcurrentDictionary<string, SemaphoreSlim>();
            _queue = new ConcurrentQueue<QueuedBuild>();
        }

        public async Task<bool> EnqueueAsync(QueuedBuild build, CancellationToken cancellationToken)
        {
            _queue.Enqueue(build);
            var customerSemaphore = _customerSemaphores.GetOrAdd(build.CustomerId, _ => new SemaphoreSlim(1, 1));
            try
            {
                await _semaphore.WaitAsync(cancellationToken);
                await customerSemaphore.WaitAsync(cancellationToken);
                return true;
            }
            catch (OperationCanceledException)
            {
                return false;
            }

        }
        public void ReleaseBuild(string customerId)
        {
            if (_customerSemaphores.TryGetValue(customerId, out var customerSemaphore))
            {
                customerSemaphore.Release();
            }
            _semaphore.Release();
        }

        public int GetQueueDepth()
        {
            return _queue.Count;
        }
        public int GetAvailableSlots()
        {
            return _semaphore.CurrentCount;
        }
    }


}
/// <summary>
/// Represents a build queued for execution.
/// </summary>
public record QueuedBuild(string BuildId, string CustomerId, string Target, string ConfigPath, DateTime QueuedAt);
