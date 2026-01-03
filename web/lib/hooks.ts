"use client";

import { useState, useEffect } from "react";
import { apiClient, Agent, ListAgentsResponse } from "./api";

export function useAgents() {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchAgents = async () => {
      try {
        setLoading(true);
        const data = await apiClient.listAgents();
        setAgents(data.agents || []);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err : new Error("Unknown error"));
        setAgents([]);
      } finally {
        setLoading(false);
      }
    };

    fetchAgents();
    // Poll for updates every 10 seconds
    const interval = setInterval(fetchAgents, 10000);
    return () => clearInterval(interval);
  }, []);

  return { agents, loading, error };
}

export function useAgent(agentID: string) {
  const [agent, setAgent] = useState<Agent | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!agentID) return;

    const fetchAgent = async () => {
      try {
        setLoading(true);
        const data = await apiClient.getAgent(agentID);
        setAgent(data);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err : new Error("Unknown error"));
        setAgent(null);
      } finally {
        setLoading(false);
      }
    };

    fetchAgent();
    const interval = setInterval(fetchAgent, 10000);
    return () => clearInterval(interval);
  }, [agentID]);

  return { agent, loading, error };
}
