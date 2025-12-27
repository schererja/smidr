# Job Controller Responsibilities

## Purpose

The Job Controller is a serverless Lambda function responsible for orchestrating jobs between the Control Plane and agents via SQS. It ensures that jobs are queued, assigned, and executed according to policies and tenant constraints.

---

## Responsibilities

- **Job Queuing:** Place jobs on SQS queue after Control Plane validation
- **Job Matching:** Select agents with compatible capabilities from DynamoDB registry
- **Policy Enforcement:** Apply execution constraints such as resource limits, approved plugins, or execution windows
- **State Tracking:** Monitor job progress through the lifecycle (Submitted → Queued → Dispatched → Running → Completed) in DynamoDB
- **Retries and Failures:** Handle failed, timed out, or cancelled jobs according to policy
- **Reporting:** Provide logs, status updates, and metrics to the Control Plane and frontend via DynamoDB
- **Integration:** Communicate with artifact store (S3/DynamoDB) and policy engine for complete lifecycle visibility

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
      | (Policy, DynamoDB)     |
      +-----------+------------+
                  |
                  v
         +-----------------+
         | Job Controller  |
         | (Lambda)        |
         +---+---------+---+
             |         |
   Queue job |         |Monitor state
        on SQS|         |in DynamoDB
             v         v
       +-----------+  +-----------+
       |  Agent 1  |  |  Agent N  |
       +-----------+  +-----------+
             |         |
       Reports via REST+HMAC
             v
      +-----------------+
      | Control Plane   |
      | (State store)   |
      +-----------------+
```
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
