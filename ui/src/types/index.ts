export type HealthState = 'learning' | 'healthy' | 'degraded' | 'attention' | 'unknown';

export interface Agent {
  agentId: string;
  hostname: string;
  lastHeartbeat: string;
  healthState: HealthState;
  registeredAt: string;
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
}

export interface AgentDetail extends Agent {
  baselines: Baseline[];
  recentHeartbeats: Heartbeat[];
}
