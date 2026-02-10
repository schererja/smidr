# Yggdrasil Licensing Guide

This document explains the licensing structure for the Yggdrasil MSP/CRM/ERP platform and how it applies to different components.

## Overview

Yggdrasil uses a **mixed licensing approach** to balance open source collaboration with sustainable business development:

- **Core Platform**: Business Source License 1.1 (BUSL)
- **Plugins & SDKs**: MIT License

## Licensing Structure

### BUSL 1.1 - Core Platform

The core platform components are licensed under the Business Source License 1.1 to protect our investment while enabling community development.

**Covered Components:**

```bash
cmd/                  # Entry points (api, agent)
internal/             # Private application code
├── ticket/           # Domain packages
├── client/
├── agent/
├── auth/
├── platform/         # Infrastructure
└── shared/           # Shared utilities
```

**Key Restrictions:**

- Cannot sell access to the Licensed Work as a service
- Cannot redistribute the Licensed Work
- Cannot use the Licensed Work for competitive purposes

**Allowed Uses:**

- Internal use within your organization
- Development and testing
- Contributing to the project

**Change Date**: January 1, 2029
After this date, BUSL components will convert to MIT License permanently.

### MIT License - Plugins & SDKs

Components designed for extension and integration are licensed under MIT to encourage community contributions and third-party development.

**Covered Components:**

```bash
packages/             # SDKs, proto definitions, shared utilities
├── agent-sdk/        # Agent plugin SDK
├── proto/            # Protocol buffer definitions
└── shared/           # Shared code

web/                  # React TypeScript frontend
├── src/
├── public/
└── components/
```

**Allowed Uses:**

- Commercial use
- Distribution
- Modification
- Private use
- Sublicensing

## Usage Scenarios

### ✅ Permitted Under Both Licenses

1. **Internal Deployment**

   ```bash
   # Deploy Yggdrasil within your company
   docker compose up -d
   ```

2. **Custom Plugin Development**

   ```typescript
   // Use MIT-licensed SDK to build custom plugins
   import { PluginSDK } from "@yggdrasil/sdk";
   ```

3. **Contributing to Core Platform**

   ```bash
   # Submit PRs to BUSL components
   git push origin feature/new-feature
   ```

### ❌ Restricted Under BUSL

1. **SaaS Competition**
   - Cannot offer Yggdrasil as a competing service
   - Cannot resell access to the core platform

2. **Redistribution of Core**
   - Cannot distribute control-plane or agent code
   - Cannot fork and redistribute core components

### ✅ Permitted Under MIT

1. **Plugin Distribution**

   ```typescript
   // Distribute your custom plugins
   npm publish my-yggdrasil-plugin
   ```

2. **Frontend Customization**

   ```bash
   # Fork and customize the web interface
   git clone https://github.com/your-org/yggdrasil-web
   ```

3. **SDK Integration**

   ```go
   // Use Go SDK in commercial applications
   import "github.com/intrik8-labs/yggdrasil-go-sdk"
   ```

## License Conversion

On January 1, 2029, the BUSL components will automatically convert to MIT License. This means:

- All restrictions will be lifted
- Core platform becomes fully open source
- Existing users gain full redistribution rights

## Compliance Guidelines

### For Internal Use

No special licensing required. Both BUSL and MIT components can be used internally without restrictions.

### For Plugin Development

1. **Use MIT Components**: Leverage packages/ and web/ freely
2. **Interact with Core**: Use APIs and SDKs to extend functionality
3. **Distribution**: Your plugins can be commercially licensed

### For Service Providers

1. **Cannot Compete**: Do not offer Yggdrasil as a competing service
2. **Can Integrate**: Use SDKs to integrate with existing systems
3. **Can Extend**: Build value-add services on top of the platform

### For Contributors

1. **CLA Required**: All contributors must sign a Contributor License Agreement
2. **License Assignment**: Contributions follow the target component's license
3. **Commercial Contributions**: Welcomed under appropriate licensing

## Frequently Asked Questions

### Q: Can I run Yggdrasil for my MSP business?

**A**: Yes! Internal use is fully permitted under BUSL.

### Q: Can I sell a custom plugin I built?

**A**: Yes! Plugins using MIT components can be commercially licensed.

### Q: Can I fork the entire project?

**A**: You can fork MIT components but cannot redistribute BUSL components.

### Q: What happens after the Change Date?

**A**: All components become MIT licensed with full redistribution rights.

### Q: Can I contribute to BUSL components?

**A**: Yes! Contributions are welcome and follow BUSL licensing.

## License Files

- `LICENSE` - Main BUSL 1.1 license for core platform
- `packages/LICENSE-MIT` - MIT license for plugins and SDKs
- `docs/CODE_OF_CONDUCT.md` - Community guidelines
- `docs/CONTRIBUTING.md` - Contribution guidelines

## Contact

For licensing questions:

- Email: <licensing@yggdrasil.dev>
- GitHub Issues: Use the "question" template

## Legal Notice

This licensing guide is for informational purposes only. The actual license files contain the legally binding terms. For legal advice regarding your specific use case, please consult with legal counsel.

---

_This licensing approach supports sustainable open source development while encouraging community participation and commercial innovation._
