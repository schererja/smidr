import React, { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { AgentDetail } from '../types';
import { agentApi } from '../api/client';
import HealthBadge from '../components/HealthBadge';
import MetricsCard from '../components/MetricsCard';
import { Card, CardContent, CardHeader, CardTitle } from '@/ui/card';
import { ChevronRight } from 'lucide-react';

const SystemDetail: React.FC = () => {
  const { agentId } = useParams<{ agentId: string }>();
  const [agent, setAgent] = useState<AgentDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!agentId) return;

    const loadAgent = async () => {
      try {
        const data = await agentApi.getAgent(agentId);
        setAgent(data);
        setLoading(false);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load agent');
        setLoading(false);
      }
    };

    loadAgent();
    const interval = setInterval(loadAgent, 30000);
    return () => clearInterval(interval);
  }, [agentId]);

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-6">
        <div className="text-center py-12 text-gray-500">Loading agent details...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-7xl mx-auto px-6">
        <div className="bg-red-50 border border-red-200 text-red-800 rounded-lg p-4">
          Error: {error}
        </div>
      </div>
    );
  }

  if (!agent) {
    return (
      <div className="max-w-7xl mx-auto px-6">
        <div className="text-center py-12 text-gray-500">Agent not found</div>
      </div>
    );
  }

  const formatTimestamp = (timestamp: string): string => {
    return new Date(timestamp).toLocaleString();
  };

  return (
    <div className="max-w-7xl mx-auto px-6 space-y-6">
      <nav className="flex items-center text-sm text-gray-600">
        <Link to="/" className="hover:text-purple-600">Systems</Link>
        <ChevronRight className="h-4 w-4 mx-2" />
        <span className="text-gray-900 font-medium">{agent.hostname}</span>
      </nav>

      <div className="flex justify-between items-start">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">{agent.hostname}</h1>
          <p className="text-sm text-gray-600 mt-1 font-mono">Agent ID: {agent.agentId}</p>
        </div>
        <HealthBadge state={agent.healthState} size="large" />
      </div>

      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-500">Last Heartbeat</span>
            <span className="text-sm font-medium">{formatTimestamp(agent.lastHeartbeat)}</span>
          </div>
        </CardContent>
      </Card>

      <MetricsCard signals={agent.signals} baselines={agent.baselines} />

      {agent.baselines && agent.baselines.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Learned Baselines</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {agent.baselines.map((baseline) => (
                <div key={baseline.metric} className="border rounded-lg p-4">
                  <h4 className="font-semibold text-gray-900 mb-2">{baseline.metric}</h4>
                  <div className="space-y-1 text-sm">
                    <div className="flex justify-between">
                      <span className="text-gray-500">Mean:</span>
                      <span className="font-medium">{baseline.mean.toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">Std Dev:</span>
                      <span className="font-medium">{baseline.stdDev.toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">Range:</span>
                      <span className="font-medium">
                        {baseline.min.toFixed(1)} - {baseline.max.toFixed(1)}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {agent.recentHeartbeats && agent.recentHeartbeats.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Recent Heartbeats</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {agent.recentHeartbeats.map((hb, idx) => (
                <div
                  key={idx}
                  className="flex justify-between items-center py-3 border-b last:border-0"
                >
                  <span className="text-sm text-gray-600">{formatTimestamp(hb.timestamp)}</span>
                  <div className="flex gap-4 text-sm">
                    <span className="text-gray-700">Load: {hb.loadAverage1m.toFixed(2)}</span>
                    <span className="text-gray-700">Mem: {hb.memoryUsedPct.toFixed(1)}%</span>
                    <span className="text-gray-700">Disk: {hb.diskUsedPct.toFixed(1)}%</span>
                    <span className="text-gray-700">Procs: {hb.processCount}</span>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
};

export default SystemDetail;
