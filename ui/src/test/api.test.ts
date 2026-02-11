import { describe, it, expect, vi, beforeEach } from 'vitest';
import axios from 'axios';

vi.mock('axios');

describe('agentApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('listAgents', () => {
    it('should fetch and map agents correctly', async () => {
      const mockApiResponse = [
        {
          id: 'agent-1',
          hostname: 'host1',
          currentHealth: 'Healthy',
          lastHeartbeatAt: '2025-01-01T00:00:00Z',
          registeredAt: '2025-01-01T00:00:00Z',
          revokedAt: null,
          latestSignals: {
            uptimeSeconds: 1000,
            loadAverage1m: 1.5,
            memoryUsedPct: 50,
            diskUsedPct: 30,
            processCount: 100,
          },
        },
      ];

      const mockGet = vi.fn().mockResolvedValue({ data: mockApiResponse });
      vi.mocked(axios.create).mockReturnValue({ get: mockGet } as any);

      // Import after mocking
      const { agentApi } = await import('../api/client');

      const agents = await agentApi.listAgents();

      expect(agents).toHaveLength(1);
      expect(agents[0].agentId).toBe('agent-1');
      expect(agents[0].hostname).toBe('host1');
      expect(agents[0].healthState).toBe('healthy');
      expect(agents[0].signals?.uptimeSeconds).toBe(1000);
    });

    it('should handle ECONNREFUSED error', async () => {
      const mockGet = vi.fn().mockRejectedValue({
        isAxiosError: true,
        code: 'ECONNREFUSED',
      });
      vi.mocked(axios.create).mockReturnValue({ get: mockGet } as any);

      const { agentApi } = await import('../api/client');

      await expect(agentApi.listAgents()).rejects.toThrow(
        'Cannot connect to control plane API'
      );
    });

    it('should handle 500 server error', async () => {
      const mockGet = vi.fn().mockRejectedValue({
        isAxiosError: true,
        response: { status: 500 },
        message: 'Internal Server Error',
      });
      vi.mocked(axios.create).mockReturnValue({ get: mockGet } as any);

      const { agentApi } = await import('../api/client');

      await expect(agentApi.listAgents()).rejects.toThrow(
        'Server error loading agents'
      );
    });

    it('should map health states case-insensitively', async () => {
      const mockApiResponse = [
        {
          id: 'agent-1',
          hostname: 'host1',
          currentHealth: 'LEARNING',
          lastHeartbeatAt: null,
          registeredAt: '2025-01-01T00:00:00Z',
          revokedAt: null,
          latestSignals: null,
        },
        {
          id: 'agent-2',
          hostname: 'host2',
          currentHealth: 'Attention',
          lastHeartbeatAt: null,
          registeredAt: '2025-01-01T00:00:00Z',
          revokedAt: null,
          latestSignals: null,
        },
      ];

      const mockGet = vi.fn().mockResolvedValue({ data: mockApiResponse });
      vi.mocked(axios.create).mockReturnValue({ get: mockGet } as any);

      const { agentApi } = await import('../api/client');

      const agents = await agentApi.listAgents();

      expect(agents[0].healthState).toBe('learning');
      expect(agents[1].healthState).toBe('attention');
    });
  });

  describe('getAgent', () => {
    it('should fetch and map agent detail correctly', async () => {
      const mockApiResponse = {
        id: 'agent-1',
        hostname: 'host1',
        currentHealth: 'Healthy',
        lastHeartbeatAt: '2025-01-01T00:00:00Z',
        registeredAt: '2025-01-01T00:00:00Z',
        revokedAt: null,
        latestSignals: {
          uptimeSeconds: 1000,
          loadAverage1m: 1.5,
          memoryUsedPct: 50,
          diskUsedPct: 30,
          processCount: 100,
        },
        baselines: [
          {
            metricName: 'LoadAverage1m',
            mean: 1.2,
            stdDev: 0.3,
            min: 0.5,
            max: 2.0,
            sampleCount: 100,
          },
        ],
        recentHeartbeats: [
          {
            timestamp: '2025-01-01T00:00:00Z',
            uptimeSeconds: 1000,
            loadAverage1m: 1.5,
            memoryUsedPct: 50,
            diskUsedPct: 30,
            processCount: 100,
          },
        ],
      };

      const mockGet = vi.fn().mockResolvedValue({ data: mockApiResponse });
      vi.mocked(axios.create).mockReturnValue({ get: mockGet } as any);

      const { agentApi } = await import('../api/client');

      const agent = await agentApi.getAgent('agent-1');

      expect(agent.agentId).toBe('agent-1');
      expect(agent.baselines).toHaveLength(1);
      expect(agent.baselines[0].metric).toBe('LoadAverage1m');
      expect(agent.recentHeartbeats).toHaveLength(1);
      expect(agent.recentHeartbeats[0].agentId).toBe('agent-1');
    });

    it('should handle 404 not found', async () => {
      const mockGet = vi.fn().mockRejectedValue({
        isAxiosError: true,
        response: { status: 404 },
        message: 'Not Found',
      });
      vi.mocked(axios.create).mockReturnValue({ get: mockGet } as any);

      const { agentApi } = await import('../api/client');

      await expect(agentApi.getAgent('nonexistent')).rejects.toThrow(
        'Agent not found'
      );
    });

    it('should default lastHeartbeat to epoch if null', async () => {
      const mockApiResponse = {
        id: 'agent-1',
        hostname: 'host1',
        currentHealth: 'Learning',
        lastHeartbeatAt: null,
        registeredAt: '2025-01-01T00:00:00Z',
        revokedAt: null,
        latestSignals: null,
        baselines: [],
        recentHeartbeats: [],
      };

      const mockGet = vi.fn().mockResolvedValue({ data: mockApiResponse });
      vi.mocked(axios.create).mockReturnValue({ get: mockGet } as any);

      const { agentApi } = await import('../api/client');

      const agent = await agentApi.getAgent('agent-1');

      expect(agent.lastHeartbeat).toBe(new Date(0).toISOString());
    });
  });
});
