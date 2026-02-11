import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Agent } from '../types';
import { agentApi } from '../api/client';
import HealthBadge from '../components/HealthBadge';
import { Input } from '@/ui/input';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/ui/table';
import { Search } from 'lucide-react';

const SystemList: React.FC = () => {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    const loadAgents = async () => {
      try {
        const data = await agentApi.listAgents();
        setAgents(data);
        setLoading(false);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load agents');
        setLoading(false);
      }
    };

    loadAgents();
    const interval = setInterval(loadAgents, 30000);
    return () => clearInterval(interval);
  }, []);

  const formatLastSeen = (timestamp: string): string => {
    const heartbeatTime = new Date(timestamp).getTime();
    
    // Handle epoch time (never received heartbeat)
    if (heartbeatTime === 0) return 'Never';
    
    const diffMs = Date.now() - heartbeatTime;
    const seconds = Math.floor(Math.abs(diffMs) / 1000);
    
    // Handle future timestamps (clock skew)
    if (diffMs < 0) return 'Just now';
    
    if (seconds < 60) return `${seconds}s ago`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
    return `${Math.floor(seconds / 86400)}d ago`;
  };

  const formatDate = (timestamp: string): string => {
    return new Date(timestamp).toLocaleString();
  };

  const filteredAgents = agents.filter((agent) =>
    agent.hostname.toLowerCase().includes(searchTerm.toLowerCase()) ||
    agent.agentId.toLowerCase().includes(searchTerm.toLowerCase())
  );

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-6">
        <div className="text-center py-12 text-gray-500">Loading agents...</div>
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

  return (
    <div className="max-w-7xl mx-auto px-6">
      <div className="mb-6">
        <div className="flex justify-between items-center mb-4">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Systems</h1>
            <p className="text-sm text-gray-500 mt-1">{filteredAgents.length} of {agents.length} systems</p>
          </div>
        </div>
        
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
          <Input
            type="text"
            placeholder="Search by hostname or agent ID..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>
      </div>

      <div className="bg-white rounded-lg border border-gray-200 shadow-sm">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Hostname</TableHead>
              <TableHead>Agent ID</TableHead>
              <TableHead>Health State</TableHead>
              <TableHead>Load Avg</TableHead>
              <TableHead>Memory</TableHead>
              <TableHead>Disk</TableHead>
              <TableHead>Last Heartbeat</TableHead>
              <TableHead>Registered</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredAgents.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8} className="text-center py-8 text-gray-500">
                  {searchTerm ? 'No agents match your search' : 'No agents registered yet'}
                </TableCell>
              </TableRow>
            ) : (
              filteredAgents.map((agent) => (
                <TableRow
                  key={agent.agentId}
                  onClick={() => navigate(`/systems/${agent.agentId}`)}
                  className="cursor-pointer hover:bg-gray-50"
                >
                  <TableCell className="font-medium">{agent.hostname}</TableCell>
                  <TableCell className="font-mono text-xs text-gray-600">
                    {agent.agentId.slice(0, 12)}...
                  </TableCell>
                  <TableCell>
                    <HealthBadge state={agent.healthState} size="small" />
                  </TableCell>
                  <TableCell>
                    {agent.signals ? agent.signals.loadAverage1m.toFixed(2) : '—'}
                  </TableCell>
                  <TableCell>
                    {agent.signals ? `${agent.signals.memoryUsedPct.toFixed(0)}%` : '—'}
                  </TableCell>
                  <TableCell>
                    {agent.signals ? `${agent.signals.diskUsedPct.toFixed(0)}%` : '—'}
                  </TableCell>
                  <TableCell className="text-gray-600">
                    {formatLastSeen(agent.lastHeartbeat)}
                  </TableCell>
                  <TableCell className="text-gray-600 text-xs">
                    {formatDate(agent.registeredAt)}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
};

export default SystemList;
