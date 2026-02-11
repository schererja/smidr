import React, { useEffect, useState } from 'react';
import { Agent } from '../types';
import { agentApi } from '../api/client';
import StatCard from '../components/StatCard';
import AgentCard from '../components/AgentCard';
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
import { 
  Search, 
  Server, 
  CheckCircle, 
  AlertTriangle, 
  Clock,
  Activity,
  TrendingUp
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/ui/card';
import { useNavigate } from 'react-router-dom';

type ViewMode = 'grid' | 'table';

const SystemList: React.FC = () => {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [viewMode, setViewMode] = useState<ViewMode>('grid');
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

  // Calculate summary stats
  const totalAgents = agents.length;
  const healthyAgents = agents.filter(a => a.healthState === 'healthy').length;
  const issuesAgents = agents.filter(a => ['degraded', 'attention'].includes(a.healthState)).length;
  const learningAgents = agents.filter(a => a.healthState === 'learning').length;
  const avgUptime = agents.length > 0 
    ? agents.reduce((sum, a) => sum + (a.signals?.uptimeSeconds || 0), 0) / agents.length 
    : 0;
  const uptimeDays = Math.floor(avgUptime / 86400);

  if (loading) {
    return (
      <div className="p-8">
        <div className="flex items-center justify-center h-64">
          <div className="flex flex-col items-center gap-3">
            <Activity className="w-8 h-8 text-purple-600 animate-pulse" />
            <p className="text-gray-500">Loading agents...</p>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8">
        <div className="bg-red-50 border border-red-200 text-red-800 rounded-lg p-6 max-w-2xl">
          <h3 className="font-semibold mb-2">Connection Error</h3>
          <p>{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-4 md:p-8 space-y-6 md:space-y-8">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl md:text-3xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-sm md:text-base text-gray-600 mt-1">Monitor your infrastructure health and performance</p>
      </div>

      {/* Summary Stats */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 md:gap-6">
        <StatCard
          title="Total Agents"
          value={totalAgents}
          subtitle={`${learningAgents} learning`}
          icon={Server}
          colorClass="bg-purple-100 text-purple-600"
        />
        <StatCard
          title="Healthy"
          value={healthyAgents}
          subtitle={`${totalAgents > 0 ? ((healthyAgents / totalAgents) * 100).toFixed(0) : 0}% of fleet`}
          icon={CheckCircle}
          colorClass="bg-green-100 text-green-600"
        />
        <StatCard
          title="Issues"
          value={issuesAgents}
          subtitle={issuesAgents > 0 ? 'Needs attention' : 'All clear'}
          icon={AlertTriangle}
          colorClass={issuesAgents > 0 ? 'bg-red-100 text-red-600' : 'bg-gray-100 text-gray-600'}
        />
        <StatCard
          title="Avg Uptime"
          value={`${uptimeDays}d`}
          subtitle={`${Math.floor(avgUptime / 3600) % 24}h ${Math.floor(avgUptime / 60) % 60}m`}
          icon={TrendingUp}
          colorClass="bg-blue-100 text-blue-600"
        />
      </div>

      {/* Search and View Toggle */}
      <Card>
        <CardContent className="p-6">
          <div className="flex items-center gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
              <Input
                type="text"
                placeholder="Search by hostname or agent ID..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>
            <div className="flex gap-2 border border-gray-200 rounded-lg p-1">
              <button
                onClick={() => setViewMode('grid')}
                className={`px-3 py-1.5 rounded text-sm font-medium transition-colors ${
                  viewMode === 'grid'
                    ? 'bg-purple-100 text-purple-700'
                    : 'text-gray-600 hover:bg-gray-100'
                }`}
              >
                Grid
              </button>
              <button
                onClick={() => setViewMode('table')}
                className={`px-3 py-1.5 rounded text-sm font-medium transition-colors ${
                  viewMode === 'table'
                    ? 'bg-purple-100 text-purple-700'
                    : 'text-gray-600 hover:bg-gray-100'
                }`}
              >
                Table
              </button>
            </div>
          </div>
          {filteredAgents.length !== agents.length && (
            <p className="text-sm text-gray-600 mt-3">
              Showing {filteredAgents.length} of {agents.length} agents
            </p>
          )}
        </CardContent>
      </Card>

      {/* Agents Display */}
      {filteredAgents.length === 0 ? (
        <Card>
          <CardContent className="p-8 md:p-12 text-center">
            <Server className="w-10 h-10 md:w-12 md:h-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-base md:text-lg font-semibold text-gray-900 mb-2">
              {searchTerm ? 'No agents match your search' : 'No agents registered yet'}
            </h3>
            <p className="text-sm md:text-base text-gray-600">
              {searchTerm
                ? 'Try adjusting your search terms'
                : 'Register your first agent to start monitoring'}
            </p>
          </CardContent>
        </Card>
      ) : viewMode === 'grid' ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 md:gap-6">
          {filteredAgents.map((agent) => (
            <AgentCard key={agent.agentId} agent={agent} />
          ))}
        </div>
      ) : (
        <Card>
          <CardContent className="p-0 overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Hostname</TableHead>
                  <TableHead className="hidden md:table-cell">Agent ID</TableHead>
                  <TableHead>Health</TableHead>
                  <TableHead className="hidden lg:table-cell">Load</TableHead>
                  <TableHead className="hidden lg:table-cell">Memory</TableHead>
                  <TableHead className="hidden lg:table-cell">Disk</TableHead>
                  <TableHead className="hidden sm:table-cell">Last Seen</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredAgents.map((agent) => (
                  <TableRow
                    key={agent.agentId}
                    onClick={() => navigate(`/systems/${agent.agentId}`)}
                    className="cursor-pointer hover:bg-gray-50"
                  >
                    <TableCell className="font-medium">{agent.hostname}</TableCell>
                    <TableCell className="hidden md:table-cell font-mono text-xs text-gray-600">
                      {agent.agentId.slice(0, 12)}...
                    </TableCell>
                    <TableCell>
                      <HealthBadge state={agent.healthState} size="small" />
                    </TableCell>
                    <TableCell className="hidden lg:table-cell">
                      {agent.signals ? agent.signals.loadAverage1m.toFixed(2) : '—'}
                    </TableCell>
                    <TableCell className="hidden lg:table-cell">
                      {agent.signals ? `${agent.signals.memoryUsedPct.toFixed(0)}%` : '—'}
                    </TableCell>
                    <TableCell className="hidden lg:table-cell">
                      {agent.signals ? `${agent.signals.diskUsedPct.toFixed(0)}%` : '—'}
                    </TableCell>
                    <TableCell className="hidden sm:table-cell text-gray-600 text-xs md:text-sm">
                      {formatLastSeen(agent.lastHeartbeat)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </div>
  );
};

export default SystemList;
