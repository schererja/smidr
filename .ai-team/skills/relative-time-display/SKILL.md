---
name: "relative-time-display"
description: "Displaying relative timestamps with clock skew tolerance and sentinel value handling"
domain: "frontend"
confidence: "low"
source: "earned"
---

## Context
When displaying "time ago" strings in user interfaces (e.g., "5 minutes ago", "2 days ago"), naive implementations often fail in production due to clock skew between client and server, or special sentinel values like epoch time. This skill documents robust patterns for displaying relative timestamps that handle these edge cases gracefully.

## Core Pattern

### Robust Time Difference Calculation
```typescript
const formatLastSeen = (timestamp: string): string => {
  const heartbeatTime = new Date(timestamp).getTime();
  
  // Handle epoch time (never received data)
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
```

## Key Principles

### 1. Detect Sentinel Values
Check for special timestamps that indicate "no data yet":
```typescript
const heartbeatTime = new Date(timestamp).getTime();
if (heartbeatTime === 0) return 'Never';
```

Common sentinel values:
- Epoch time: `1970-01-01T00:00:00.000Z` (timestamp = 0)
- Far future: `9999-12-31T23:59:59.999Z` (max timestamp)
- Null/undefined: Handle before parsing

### 2. Use Math.abs() for Clock Skew
Always take the absolute value of the time difference:
```typescript
const seconds = Math.floor(Math.abs(diffMs) / 1000);
```

This prevents negative values from appearing in the UI.

### 3. Handle Future Timestamps Explicitly
Check if timestamp is in the future before formatting:
```typescript
const diffMs = Date.now() - heartbeatTime;
if (diffMs < 0) return 'Just now';
```

Alternatives:
- `'Just now'` — implies very recent
- `'A moment ago'` — more vague
- `'< 1s ago'` — technically accurate

### 4. Graceful Degradation
Provide sensible defaults for edge cases:
```typescript
// Invalid date string
if (isNaN(heartbeatTime)) return 'Unknown';

// Very old timestamps (> 365 days)
if (seconds > 31536000) return formatDate(timestamp); // Fall back to absolute date
```

## Common Anti-Patterns

### ❌ Naive Subtraction
```typescript
// BAD: Produces negative values with clock skew
const seconds = (Date.now() - new Date(timestamp).getTime()) / 1000;
return `${seconds}s ago`; // Shows "-5s ago"
```

### ❌ No Sentinel Handling
```typescript
// BAD: Shows "55 years ago" for never-received data
const seconds = (Date.now() - new Date(timestamp).getTime()) / 1000;
return formatSeconds(seconds); // Shows "55y ago" for epoch time
```

### ❌ Silent Failures
```typescript
// BAD: NaN propagates through UI
const seconds = (Date.now() - new Date(invalidString).getTime()) / 1000;
return `${Math.floor(seconds)}s ago`; // Shows "NaNs ago"
```

## When to Apply

Use this pattern when:
- Displaying "time ago" strings for user activity, heartbeats, or events
- Client and server clocks may not be perfectly synchronized
- Data model uses sentinel values (epoch, null) for "no data yet"
- Working with distributed systems where clock skew is common
- Users may have incorrect system time

## Real-World Example

**Symptoms:** UI showing "-5 seconds ago" for recent heartbeats

**Root Cause:** Server timestamp is 5 seconds ahead of client clock (clock skew)

**Fix:** Apply clock skew tolerance and sentinel handling as shown above

**Result:** Displays "Just now" instead of negative time

## Trade-offs

**Advantages:**
- Handles clock skew gracefully (common in distributed systems)
- Displays user-friendly text for special cases
- Prevents confusing negative time values
- Robust against invalid input

**Disadvantages:**
- Hides information about clock skew (may mask infrastructure issues)
- "Just now" is less precise than actual time difference
- Requires more code than naive subtraction

## Alternative Approaches

### Server-Side Rendering
Let the server calculate "ago" strings:
```typescript
// API response includes pre-computed string
{ lastHeartbeat: "5 minutes ago" }
```

**Pros:** Eliminates clock skew, consistent across clients  
**Cons:** Stale values, requires API changes, i18n complexity

### Relative Time Libraries
Use libraries like `date-fns` or `dayjs`:
```typescript
import { formatDistanceToNow } from 'date-fns';
formatDistanceToNow(new Date(timestamp), { addSuffix: true });
```

**Pros:** Battle-tested, i18n support, handles edge cases  
**Cons:** Additional dependency, may still need sentinel handling

## References

- Fixed in: `ui/src/pages/SystemList.tsx:41-56`
- Context: Clock skew between control plane (server) and UI (client)
- Sentinel value: `new Date(0).toISOString()` for agents with no heartbeat
