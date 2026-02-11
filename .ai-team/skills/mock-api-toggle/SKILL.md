---
name: "mock-api-toggle"
description: "Pattern for using mock data with toggle flag to enable parallel frontend/backend development"
domain: "api-integration"
confidence: "medium"
source: "earned"
---

## Context

When building a frontend that depends on a backend API that's still in development, use a mock data pattern with a toggle flag. This allows frontend and backend teams to work independently while ensuring the contract is clear.

## Patterns

1. **Single API client module** - All API calls go through one module (e.g., `api/client.ts`)
2. **Toggle flag** - Boolean constant (`USE_MOCK`) at top of API client to switch between mock and real API
3. **Same interface** - Mock data returns same TypeScript types as real API will
4. **Realistic mock data** - Include edge cases: stale data, different health states, missing fields
5. **Easy transition** - Changing flag from `true` to `false` is the only code change needed

## Examples

```typescript
// api/client.ts
import axios from 'axios';
import { Agent, AgentDetail } from '../types';

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
});

const mockAgents: Agent[] = [
  // Realistic mock data with various states
  { agentId: 'xxx', hostname: 'web-01', healthState: 'healthy', ... },
  { agentId: 'yyy', hostname: 'db-01', healthState: 'learning', ... },
  { agentId: 'zzz', hostname: 'cache-01', healthState: 'degraded', ... },
];

const USE_MOCK = true; // Toggle this when backend is ready

export const agentApi = {
  async listAgents(): Promise<Agent[]> {
    if (USE_MOCK) {
      return Promise.resolve(mockAgents);
    }
    const response = await api.get('/v0/agents');
    return response.data;
  },
  // ... other methods
};
```

## Anti-Patterns

- Don't scatter mock data across components - centralize in API client
- Don't use different types for mock vs real data - enforce same contract
- Don't forget to mock error states and edge cases
- Don't hardcode mock data in components - keep it in API layer
- Don't use environment variables for the toggle in early dev - simple constant is easier
