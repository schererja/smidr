# Ceremonies

> Team meetings that happen before or after work. Each squad configures their own.

## Design Review

| Field | Value |
|-------|-------|
| **Trigger** | auto |
| **When** | before |
| **Condition** | multi-agent task involving 2+ agents modifying shared systems (agent ↔ control plane communication, mTLS flows) |
| **Facilitator** | ripley |
| **Participants** | all-relevant |
| **Time budget** | focused |
| **Enabled** | ✅ yes |

**Agenda:**
1. Review the task and requirements
2. Agree on interfaces and contracts between components (agent ↔ control plane API, mTLS handshake, enrollment protocol)
3. Identify risks and edge cases
4. Assign action items

---

## Retrospective

| Field | Value |
|-------|-------|
| **Trigger** | manual |
| **When** | after |
| **Condition** | — |
| **Facilitator** | ripley |
| **Participants** | all-involved |
| **Time budget** | thorough |
| **Enabled** | ✅ yes |

**Agenda:**
1. What went well?
2. What didn't go well?
3. What should we change going forward?
4. Action items
