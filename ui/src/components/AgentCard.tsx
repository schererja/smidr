import React from 'react';
import { Card, CardContent } from '@/ui/card';
import { Agent } from '@/types';
import HealthBadge from './HealthBadge';
import OSIcon from './OSIcon';
import { Clock, HardDrive, Cpu } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

interface AgentCardProps {
  agent: Agent;
}

const AgentCard: React.FC<AgentCardProps> = ({ agent }) => {
  const navigate = useNavigate();

  const formatLastSeen = (timestamp: string): string => {
    const heartbeatTime = new Date(timestamp).getTime();
    if (heartbeatTime === 0) return 'Never';
    
    const diffMs = Date.now() - heartbeatTime;
    const seconds = Math.floor(Math.abs(diffMs) / 1000);
    
    if (diffMs < 0) return 'Just now';
    if (seconds < 60) return `${seconds}s ago`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
    return `${Math.floor(seconds / 86400)}d ago`;
  };

  return (
    <Card 
      className="hover:shadow-lg transition-all cursor-pointer border border-gray-200 hover:border-purple-300"
      onClick={() => navigate(`/systems/${agent.agentId}`)}
    >
      <CardContent className="p-6">
        <div className="flex items-start justify-between mb-4">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-gradient-to-br from-purple-500 to-purple-700 rounded-lg flex items-center justify-center">
              <OSIcon os={agent.os} className="w-5 h-5 text-white" />
            </div>
            <div>
              <h3 className="font-semibold text-gray-900">{agent.hostname}</h3>
              <p className="text-xs text-gray-500 font-mono">{agent.agentId.slice(0, 12)}...</p>
            </div>
          </div>
          <HealthBadge state={agent.healthState} size="small" />
        </div>

        <div className="grid grid-cols-2 gap-3 text-sm">
          <div className="flex items-center gap-2 text-gray-600">
            <Cpu className="w-4 h-4" />
            <span>{agent.signals?.loadAverage1m.toFixed(2) || '—'}</span>
          </div>
          <div className="flex items-center gap-2 text-gray-600">
            <HardDrive className="w-4 h-4" />
            <span>{agent.signals ? `${agent.signals.memoryUsedPct.toFixed(0)}%` : '—'}</span>
          </div>
          <div className="flex items-center gap-2 text-gray-600 col-span-2">
            <Clock className="w-4 h-4" />
            <span className="text-xs">{formatLastSeen(agent.lastHeartbeat)}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default AgentCard;
