export type HealthState = 'learning' | 'healthy' | 'degraded' | 'attention' | 'unknown';
export type OS = 'linux' | 'windows' | 'darwin' | 'unknown';

export interface Agent {
  agentId: string;
  hostname: string;
  lastHeartbeat: string;
  healthState: HealthState;
  registeredAt: string;
  os?: OS; // Optional for now, will be added by agent in future
  signals?: Signals;
}

export interface Signals {
  uptimeSeconds: number;
  loadAverage1m: number;
  memoryUsedPct: number;
  diskUsedPct: number;
  processCount: number;
}

export interface Heartbeat {
  agentId: string;
  timestamp: string;
  uptimeSeconds: number;
  loadAverage1m: number;
  memoryUsedPct: number;
  diskUsedPct: number;
  processCount: number;
}

export interface Baseline {
  metric: string;
  mean: number;
  stdDev: number;
  min: number;
  max: number;
  sampleCount: number;
}

export interface AgentDetail extends Agent {
  baselines: Baseline[];
  recentHeartbeats: Heartbeat[];
}
