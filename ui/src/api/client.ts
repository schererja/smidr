import axios, { AxiosError } from 'axios';
import { Agent, AgentDetail, Baseline, Heartbeat, HealthState } from '../types';

// API client with configurable base URL
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'https://localhost:5001';

const api = axios.create({
  baseURL: `${API_BASE_URL}/api`,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Response DTOs from control plane (camelCase from C# controller)
interface SignalValuesDto {
  uptimeSeconds: number;
  loadAverage1m: number;
  memoryUsedPct: number;
  diskUsedPct: number;
  processCount: number;
}

interface ApiAgentResponse {
  id: string;  // C# Id → camelCase id
  hostname: string;
  currentHealth: string;
  lastHeartbeatAt: string | null;
  registeredAt: string;
  revokedAt: string | null;
  latestSignals: SignalValuesDto | null;
}

interface ApiAgentDetailResponse {
  id: string;  // C# Id → camelCase id
  hostname: string;
  currentHealth: string;
  lastHeartbeatAt: string | null;
  registeredAt: string;
  revokedAt: string | null;
  latestSignals: SignalValuesDto | null;
  baselines: Array<{
    metricName: string;
    mean: number;
    stdDev: number;
    min: number;
    max: number;
    sampleCount: number;
  }>;
  recentHeartbeats: Array<{
    timestamp: string;
    uptimeSeconds: number;
    loadAverage1m: number;
    memoryUsedPct: number;
    diskUsedPct: number;
    processCount: number;
  }>;
}

// Map API health status to UI format (handle case variations)
function mapHealthState(apiHealth: string): HealthState {
  const normalized = apiHealth.toLowerCase();
  switch (normalized) {
    case 'learning':
      return 'learning';
    case 'healthy':
      return 'healthy';
    case 'degraded':
      return 'degraded';
    case 'attention':
      return 'attention';
    default:
      return 'unknown';
  }
}

// Map API agent response to UI Agent type
function mapAgent(apiAgent: ApiAgentResponse): Agent {
  return {
    agentId: apiAgent.id,
    hostname: apiAgent.hostname,
    lastHeartbeat: apiAgent.lastHeartbeatAt || new Date(0).toISOString(),
    healthState: mapHealthState(apiAgent.currentHealth),
    registeredAt: apiAgent.registeredAt,
    signals: apiAgent.latestSignals || undefined,
  };
}

// Map API baseline to UI Baseline type
function mapBaseline(apiBaseline: { metricName: string; mean: number; stdDev: number; min: number; max: number; sampleCount?: number }): Baseline {
  return {
    metric: apiBaseline.metricName,
    mean: apiBaseline.mean,
    stdDev: apiBaseline.stdDev,
    min: apiBaseline.min,
    max: apiBaseline.max,
  };
}

// Map API heartbeat to UI Heartbeat type
function mapHeartbeat(apiHeartbeat: { timestamp: string; uptimeSeconds: number; loadAverage1m: number; memoryUsedPct: number; diskUsedPct: number; processCount: number }, agentId: string): Heartbeat {
  return {
    agentId,
    timestamp: apiHeartbeat.timestamp,
    uptimeSeconds: apiHeartbeat.uptimeSeconds,
    loadAverage1m: apiHeartbeat.loadAverage1m,
    memoryUsedPct: apiHeartbeat.memoryUsedPct,
    diskUsedPct: apiHeartbeat.diskUsedPct,
    processCount: apiHeartbeat.processCount,
  };
}

export const agentApi = {
  async listAgents(): Promise<Agent[]> {
    try {
      const response = await api.get<ApiAgentResponse[]>('/agents');
      return response.data.map(mapAgent);
    } catch (error) {
      if (error instanceof AxiosError) {
        if (error.code === 'ECONNREFUSED') {
          throw new Error('Cannot connect to control plane API. Is the server running?');
        }
        if (error.response?.status === 500) {
          throw new Error('Server error loading agents');
        }
        throw new Error(`Failed to load agents: ${error.message}`);
      }
      throw error;
    }
  },

  async getAgent(agentId: string): Promise<AgentDetail> {
    try {
      const response = await api.get<ApiAgentDetailResponse>(`/agents/${agentId}`);
      const apiAgent = response.data;
      
      return {
        agentId: apiAgent.id,
        hostname: apiAgent.hostname,
        lastHeartbeat: apiAgent.lastHeartbeatAt || new Date(0).toISOString(),
        healthState: mapHealthState(apiAgent.currentHealth),
        registeredAt: apiAgent.registeredAt,
        signals: apiAgent.latestSignals || undefined,
        baselines: apiAgent.baselines.map(mapBaseline),
        recentHeartbeats: apiAgent.recentHeartbeats.map(hb => mapHeartbeat(hb, agentId)),
      };
    } catch (error) {
      if (error instanceof AxiosError) {
        if (error.code === 'ECONNREFUSED') {
          throw new Error('Cannot connect to control plane API. Is the server running?');
        }
        if (error.response?.status === 404) {
          throw new Error('Agent not found');
        }
        if (error.response?.status === 500) {
          throw new Error('Server error loading agent details');
        }
        throw new Error(`Failed to load agent: ${error.message}`);
      }
      throw error;
    }
  },
};
