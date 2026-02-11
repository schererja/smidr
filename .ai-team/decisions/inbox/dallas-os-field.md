### 2026-02-11: Added OS field to Agent model and API

**By:** Dallas

**What:** Added optional OS field to Agent entity, database schema, registration endpoint, and API responses. Agents can now report their operating system (e.g., "linux", "windows", "darwin") during registration.

**Why:** UI requires OS information to display appropriate icons for each agent. The OS field flows from agent → control plane → UI:
- Kane's agent collects and sends OS during registration
- Control plane stores OS in database and returns it in API responses
- Lambert's UI displays OS-specific icons based on this data

**Implementation:**
- Added `string? OS` property to Agent model
- Created migration 20260211000000_AddOSToAgent to add OS column
- Updated RegisterAgentRequest to accept optional OS parameter
- Included OS in both AgentListDto and AgentDetailDto responses
- Nullable field ensures backward compatibility with existing agents

**Migration applied:** Existing agents without OS data will have null values, allowing graceful degradation in the UI.