import React from 'react';
import { Signals, Baseline } from '../types';
import { Card, CardContent, CardHeader, CardTitle } from '@/ui/card';

interface MetricsCardProps {
  signals?: Signals;
  baselines?: Baseline[];
}

const MetricsCard: React.FC<MetricsCardProps> = ({ signals, baselines }) => {
  if (!signals) {
    return (
      <Card>
        <CardContent className="pt-6">
          <p className="text-gray-500">No signal data available</p>
        </CardContent>
      </Card>
    );
  }

  const formatUptime = (seconds: number): string => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    return `${days}d ${hours}h ${mins}m`;
  };

  const getBaseline = (metric: string): Baseline | undefined => {
    return baselines?.find((b) => b.metric === metric);
  };

  const formatDelta = (current: number, baseline?: Baseline): React.ReactElement | null => {
    if (!baseline) return null;
    const delta = current - baseline.mean;
    const sign = delta > 0 ? '+' : '';
    const colorClass = Math.abs(delta) > baseline.stdDev * 2 ? 'text-orange-600' : 'text-gray-500';
    return (
      <span className={`text-xs ${colorClass} ml-2`}>
        ({sign}
        {delta.toFixed(1)} from baseline)
      </span>
    );
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Current Signals</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div className="space-y-1">
            <p className="text-sm text-gray-500">Uptime</p>
            <p className="text-lg font-semibold">{formatUptime(signals.uptimeSeconds)}</p>
          </div>
          
          <div className="space-y-1">
            <p className="text-sm text-gray-500">Load Average (1m)</p>
            <p className="text-lg font-semibold">
              {signals.loadAverage1m.toFixed(2)}
              {formatDelta(signals.loadAverage1m, getBaseline('loadAverage1m'))}
            </p>
          </div>
          
          <div className="space-y-1">
            <p className="text-sm text-gray-500">Memory Used</p>
            <p className="text-lg font-semibold">
              {signals.memoryUsedPct.toFixed(1)}%
              {formatDelta(signals.memoryUsedPct, getBaseline('memoryUsedPct'))}
            </p>
          </div>
          
          <div className="space-y-1">
            <p className="text-sm text-gray-500">Disk Used</p>
            <p className="text-lg font-semibold">
              {signals.diskUsedPct.toFixed(1)}%
              {formatDelta(signals.diskUsedPct, getBaseline('diskUsedPct'))}
            </p>
          </div>
          
          <div className="space-y-1">
            <p className="text-sm text-gray-500">Process Count</p>
            <p className="text-lg font-semibold">
              {signals.processCount}
              {formatDelta(signals.processCount, getBaseline('processCount'))}
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default MetricsCard;
