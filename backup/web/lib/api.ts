// API configuration
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface Agent {
  id: string;
  name: string;
  capabilities: string[];
  metadata: Record<string, string>;
  status: string;
  registered_at?: string;
  last_heartbeat?: string;
}

export interface Job {
  id: string;
  name: string;
  status: string;
  created_at?: string;
  updated_at?: string;
}

export interface ListAgentsResponse {
  agents: Agent[];
  count: number;
}

export interface ListJobsResponse {
  jobs: Job[];
  count: number;
}

// API Client
export const apiClient = {
  // Health checks
  async healthCheck(): Promise<{ status: string; server: string }> {
    const response = await fetch(`${API_BASE_URL}/health`);
    if (!response.ok) throw new Error("Health check failed");
    return response.json();
  },

  // Agents
  async listAgents(): Promise<ListAgentsResponse> {
    const response = await fetch(`${API_BASE_URL}/api/v1/agents`);
    if (!response.ok) throw new Error("Failed to fetch agents");
    return response.json();
  },

  async getAgent(agentID: string): Promise<Agent> {
    const response = await fetch(`${API_BASE_URL}/api/v1/agents/${agentID}`);
    if (!response.ok) throw new Error("Failed to fetch agent");
    return response.json();
  },

  // Jobs
  async listJobs(): Promise<ListJobsResponse> {
    const response = await fetch(`${API_BASE_URL}/api/v1/jobs`);
    if (!response.ok) throw new Error("Failed to fetch jobs");
    return response.json();
  },

  async createJob(data: { name: string }): Promise<Job> {
    const response = await fetch(`${API_BASE_URL}/api/v1/jobs`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    });
    if (!response.ok) throw new Error("Failed to create job");
    return response.json();
  },

  async getJob(jobID: string): Promise<Job> {
    const response = await fetch(`${API_BASE_URL}/api/v1/jobs/${jobID}`);
    if (!response.ok) throw new Error("Failed to fetch job");
    return response.json();
  },
};
