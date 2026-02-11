### 2026-02-10: Documentation structure and standards

**By:** Brett

**What:** Created centralized documentation structure with three core documents (ARCHITECTURE.md, API.md, DEVELOPMENT.md) in `docs/` directory at repository root. Updated README.md to serve as entry point with links to detailed documentation.

**Why:** Establishes a clear documentation hierarchy that scales as the project grows. Separating architecture, API reference, and development guide allows different audiences (users, integrators, contributors) to find relevant information quickly. Centralizing docs in `docs/` directory follows industry convention and makes documentation discoverable. This structure prevents documentation sprawl and ensures all team members know where to add new docs as features are delivered.

**Standards Established:**
- Use markdown for all documentation
- Include runnable code examples
- Document all API endpoints with full request/response schemas
- Provide troubleshooting sections for common issues
- Cross-link between related documents
- Keep README.md concise with links to detailed docs
