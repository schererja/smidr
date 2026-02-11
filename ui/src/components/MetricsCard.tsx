import React from 'react';
import { Signals, Baseline } from '../types';
import { Card, CardContent, CardHeader, CardTitle } from '@/ui/card';
import { Activity, Cpu, HardDrive, Clock, Layers } from 'lucide-react';
import { cn } from '@/lib/utils';

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

  const formatDelta = (current: number, baseline?: Baseline): { 
    text: string; 
    severity: 'normal' | 'warning' | 'critical' 
  } | null => {
    if (!baseline) return null;
    const delta = current - baseline.mean;
    const sign = delta > 0 ? '+' : '';
    const threshold = baseline.stdDev * 2;
    const severity = Math.abs(delta) > threshold * 1.5 
      ? 'critical' 
      : Math.abs(delta) > threshold 
        ? 'warning' 
        : 'normal';
    
    return {
      text: `${sign}${delta.toFixed(1)} from baseline`,
      severity
    };
  };

  const metrics = [
    {
      label: 'Uptime',
      value: formatUptime(signals.uptimeSeconds),
      icon: Clock,
      colorClass: 'bg-blue-50 text-blue-600',
      baseline: null,
    },
    {
      label: 'Load Average (1m)',
      value: signals.loadAverage1m.toFixed(2),
      icon: Activity,
      colorClass: 'bg-purple-50 text-purple-600',
      baseline: getBaseline('loadAverage1m'),
      current: signals.loadAverage1m,
    },
    {
      label: 'Memory Used',
      value: `${signals.memoryUsedPct.toFixed(1)}%`,
      icon: Cpu,
      colorClass: 'bg-orange-50 text-orange-600',
      baseline: getBaseline('memoryUsedPct'),
      current: signals.memoryUsedPct,
    },
    {
      label: 'Disk Used',
      value: `${signals.diskUsedPct.toFixed(1)}%`,
      icon: HardDrive,
      colorClass: 'bg-green-50 text-green-600',
      baseline: getBaseline('diskUsedPct'),
      current: signals.diskUsedPct,
      // TODO: Support multiple drives when agent sends per-drive metrics
      // Currently shows aggregate disk usage for root filesystem
    },
    {
      label: 'Process Count',
      value: signals.processCount.toString(),
      icon: Layers,
      colorClass: 'bg-indigo-50 text-indigo-600',
      baseline: getBaseline('processCount'),
      current: signals.processCount,
    },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Activity className="w-5 h-5 text-purple-600" />
          Current Metrics
        </CardTitle>
        <p className="text-sm text-gray-600 mt-1">
          Real-time system performance indicators
        </p>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {metrics.map((metric) => {
            const Icon = metric.icon;
            const delta = metric.current !== undefined && metric.baseline 
              ? formatDelta(metric.current, metric.baseline) 
              : null;
            
            const deltaColorClass = delta 
              ? delta.severity === 'critical' 
                ? 'text-red-600 bg-red-50' 
                : delta.severity === 'warning' 
                  ? 'text-orange-600 bg-orange-50' 
                  : 'text-green-600 bg-green-50'
              : '';

            return (
              <Card key={metric.label} className="border-2">
                <CardContent className="p-5">
                  <div className="flex items-start justify-between mb-3">
                    <p className="text-sm font-medium text-gray-600">{metric.label}</p>
                    <div className={cn('w-10 h-10 rounded-lg flex items-center justify-center', metric.colorClass)}>
                      <Icon className="w-5 h-5" />
                    </div>
                  </div>
                  <p className="text-2xl font-bold text-gray-900 mb-2">{metric.value}</p>
                  {delta && (
                    <div className={cn('inline-flex items-center px-2 py-1 rounded text-xs font-medium', deltaColorClass)}>
                      {delta.text}
                    </div>
                  )}
                </CardContent>
              </Card>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
};

export default MetricsCard;
