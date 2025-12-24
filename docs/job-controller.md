# Job Controller Responsibilities

## Purpose

The Job Controller is responsible for orchestrating jobs between the Control Plane and agents. It ensures that jobs are assigned, tracked, and executed according to policies and tenant constraints.

---

## Responsibilities

- **Job Matching:** Assign jobs to compatible agents based on capabilities, availability, and tenant/project policies.
- **Policy Enforcement:** Apply execution constraints such as resource limits, approved plugins, or execution windows.
- **State Tracking:** Monitor job progress through the lifecycle (Submitted → Queued → Dispatched → Running → Completed).
- **Retries and Failures:** Handle failed, timed out, or cancelled jobs according to policy.
- **Reporting:** Provide logs, status updates, and metrics to the Control Plane and frontend.
- **Integration:** Communicate with artifact store and policy engine for complete lifecycle visibility.

---

## Non-Responsibilities

- Direct execution of jobs (handled by agents)
- Artifact storage or manipulation
- Agent installation or sandbox management
- Cross-tenant job awareness (enforced by the Control Plane)

---

## Job Controller Architecture

```text
      +------------------------+
      |     Control Plane      |
      | (Policy Engine, UI API)|
      +-----------+------------+
                  |
                  v
         +-----------------+
         | Job Controller  |
         +---+---------+---+
             |         |
   Assigns job|         |Monitors state
             v         v
       +-----------+  +-----------+
       |  Agent 1  |  |  Agent N  |
       +-----------+  +-----------+
             |         |
       Reports logs & status
             v
      +-----------------+
      | Artifact Store  |
      +-----------------+
```

## Job Metadata Managed by Job Controller

- Job ID

- Agent assignment

- State transitions

- Retry attempts

- Timestamped logs
- Policy decisions applied

- Error/failure codes

## User Stories

### Embedded Systems Engineer

- I want my jobs dispatched to agents that support my target hardware.

- I want consistent handling of job retries when failures occur.

### MSP Operator

- I want clear visibility of all job states across multiple tenants.

- I want jobs automatically reassigned when an agent becomes unavailable.

### Platform Administrator

- I want alerts for jobs that are stuck or failed repeatedly.

- I want an auditable trail of job dispatch decisions.

---

## Notes

- The Job Controller is the orchestrator, not the executor.

- Its design ensures multi-tenant isolation, deterministic execution, and traceable auditing.

- The Job Controller interacts closely with the Control Plane, agents, and artifact store but does not directly manipulate job outputs.
