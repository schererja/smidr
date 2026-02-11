import React, { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { AgentDetail } from '../types';
import { agentApi } from '../api/client';
import HealthBadge from '../components/HealthBadge';
import MetricsCard from '../components/MetricsCard';
import OSIcon from '../components/OSIcon';
import { Card, CardContent, CardHeader, CardTitle } from '@/ui/card';
import { 
  ChevronRight, 
  Clock, 
  TrendingUp, 
  TrendingDown,
  Minus,
  Activity
} from 'lucide-react';

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

  const getHealthColor = (state: string) => {
    switch (state) {
      case 'healthy': return 'text-green-600 bg-green-50';
      case 'degraded': return 'text-orange-600 bg-orange-50';
      case 'attention': return 'text-red-600 bg-red-50';
      case 'learning': return 'text-blue-600 bg-blue-50';
      default: return 'text-gray-600 bg-gray-50';
    }
  };

  const getStatusIcon = (current: number, baseline?: { mean: number; stdDev: number }) => {
    if (!baseline) return <Minus className="w-4 h-4 text-gray-400" />;
    const delta = current - baseline.mean;
    const threshold = baseline.stdDev * 2;
    
    if (Math.abs(delta) < threshold) {
      return <Minus className="w-4 h-4 text-green-600" />;
    }
    return delta > 0 
      ? <TrendingUp className="w-4 h-4 text-orange-600" />
      : <TrendingDown className="w-4 h-4 text-blue-600" />;
  };

  if (loading) {
    return (
      <div className="p-8">
        <div className="flex items-center justify-center h-64">
          <div className="flex flex-col items-center gap-3">
            <Activity className="w-8 h-8 text-purple-600 animate-pulse" />
            <p className="text-gray-500">Loading agent details...</p>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8">
        <div className="bg-red-50 border border-red-200 text-red-800 rounded-lg p-6 max-w-2xl">
          <h3 className="font-semibold mb-2">Error Loading Agent</h3>
          <p>{error}</p>
        </div>
      </div>
    );
  }

  if (!agent) {
    return (
      <div className="p-8">
        <Card>
          <CardContent className="p-12 text-center">
            <Activity className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Agent not found</h3>
            <p className="text-gray-600">This agent may have been removed or never existed.</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="p-4 md:p-8 space-y-4 md:space-y-6">
      {/* Breadcrumb */}
      <nav className="flex items-center text-sm text-gray-600">
        <Link to="/" className="hover:text-purple-600 transition-colors">Dashboard</Link>
        <ChevronRight className="h-4 w-4 mx-2" />
        <span className="text-gray-900 font-medium truncate">{agent.hostname}</span>
      </nav>

      {/* Hero Section */}
      <Card>
        <CardContent className="p-4 md:p-8">
          <div className="flex flex-col sm:flex-row items-start justify-between mb-4 md:mb-6 gap-4">
            <div className="flex items-center gap-3 md:gap-4">
              <div className="w-12 h-12 md:w-16 md:h-16 bg-gradient-to-br from-purple-500 to-purple-700 rounded-xl flex items-center justify-center">
                <OSIcon os={agent.os} className="w-6 h-6 md:w-8 md:h-8 text-white" />
              </div>
              <div>
                <h1 className="text-xl md:text-3xl font-bold text-gray-900">{agent.hostname}</h1>
                <p className="text-xs md:text-sm text-gray-600 mt-1 font-mono break-all">
                  Agent ID: {agent.agentId}
                </p>
              </div>
            </div>
            <HealthBadge state={agent.healthState} size="large" />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-3 md:gap-4 pt-4 md:pt-6 border-t border-gray-200">
            <div className="flex items-center gap-3">
              <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${getHealthColor(agent.healthState)}`}>
                <Clock className="w-5 h-5" />
              </div>
              <div>
                <p className="text-xs text-gray-500">Last Heartbeat</p>
                <p className="text-sm font-semibold text-gray-900">
                  {formatTimestamp(agent.lastHeartbeat)}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-blue-50 text-blue-600 rounded-lg flex items-center justify-center">
                <Activity className="w-5 h-5" />
              </div>
              <div>
                <p className="text-xs text-gray-500">Registered</p>
                <p className="text-sm font-semibold text-gray-900">
                  {formatTimestamp(agent.registeredAt)}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-purple-50 text-purple-600 rounded-lg flex items-center justify-center">
                <TrendingUp className="w-5 h-5" />
              </div>
              <div>
                <p className="text-xs text-gray-500">Baselines</p>
                <p className="text-sm font-semibold text-gray-900">
                  {agent.baselines?.length || 0} metrics tracked
                </p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Current Metrics */}
      <MetricsCard signals={agent.signals} baselines={agent.baselines} />

      {/* Baselines Detail */}
      {agent.baselines && agent.baselines.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base md:text-lg">
              <TrendingUp className="w-4 h-4 md:w-5 md:h-5 text-purple-600" />
              Learned Baselines
            </CardTitle>
            <p className="text-xs md:text-sm text-gray-600 mt-1">
              Statistical baselines computed from historical data
            </p>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 md:gap-4">
              {agent.baselines.map((baseline) => {
                const currentValue = agent.signals ? {
                  loadAverage1m: agent.signals.loadAverage1m,
                  memoryUsedPct: agent.signals.memoryUsedPct,
                  diskUsedPct: agent.signals.diskUsedPct,
                  processCount: agent.signals.processCount,
                }[baseline.metric] : undefined;

                return (
                  <Card key={baseline.metric} className="border-2">
                    <CardContent className="p-4">
                      <div className="flex items-center justify-between mb-3">
                        <h4 className="font-semibold text-gray-900 text-sm">{baseline.metric}</h4>
                        {currentValue !== undefined && getStatusIcon(currentValue, baseline)}
                      </div>
                      <div className="space-y-2 text-sm">
                        <div className="flex justify-between">
                          <span className="text-gray-600">Mean:</span>
                          <span className="font-semibold">{baseline.mean.toFixed(2)}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-gray-600">Std Dev:</span>
                          <span className="font-semibold">{baseline.stdDev.toFixed(2)}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-gray-600">Range:</span>
                          <span className="font-semibold text-xs">
                            {baseline.min.toFixed(1)} - {baseline.max.toFixed(1)}
                          </span>
                        </div>
                        <div className="flex justify-between pt-2 border-t">
                          <span className="text-gray-600">Samples:</span>
                          <span className="font-semibold">{baseline.sampleCount}</span>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                );
              })}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Recent Heartbeats */}
      {agent.recentHeartbeats && agent.recentHeartbeats.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Clock className="w-5 h-5 text-purple-600" />
              Recent Heartbeats
            </CardTitle>
            <p className="text-sm text-gray-600 mt-1">
              Latest {agent.recentHeartbeats.length} heartbeat events
            </p>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {agent.recentHeartbeats.map((hb, idx) => (
                <div
                  key={idx}
                  className="flex items-center justify-between p-4 border border-gray-200 rounded-lg hover:border-purple-300 hover:shadow-sm transition-all"
                >
                  <div className="flex items-center gap-4">
                    <div className="w-10 h-10 bg-purple-50 rounded-lg flex items-center justify-center">
                      <Activity className="w-5 h-5 text-purple-600" />
                    </div>
                    <span className="text-sm font-medium text-gray-900">
                      {formatTimestamp(hb.timestamp)}
                    </span>
                  </div>
                  <div className="flex gap-6 text-sm">
                    <div className="text-center">
                      <p className="text-xs text-gray-500 mb-1">Load</p>
                      <p className="font-semibold text-gray-900">{hb.loadAverage1m.toFixed(2)}</p>
                    </div>
                    <div className="text-center">
                      <p className="text-xs text-gray-500 mb-1">Memory</p>
                      <p className="font-semibold text-gray-900">{hb.memoryUsedPct.toFixed(1)}%</p>
                    </div>
                    <div className="text-center">
                      <p className="text-xs text-gray-500 mb-1">Disk</p>
                      <p className="font-semibold text-gray-900">{hb.diskUsedPct.toFixed(1)}%</p>
                    </div>
                    <div className="text-center">
                      <p className="text-xs text-gray-500 mb-1">Processes</p>
                      <p className="font-semibold text-gray-900">{hb.processCount}</p>
                    </div>
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
