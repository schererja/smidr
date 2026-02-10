# Administrator Guide

This comprehensive guide helps both business administrators and system administrators effectively manage the Yggdrasil MSP/CRM/ERP platform, from basic setup to advanced operations.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Understanding Your Role](#understanding-your-role)
3. [Dashboard Overview](#dashboard-overview)
4. [User Management](#user-management)
5. [Customer Management](#customer-management)
6. [Ticket Management](#ticket-management)
7. [System & Agent Management](#system--agent-management)
8. [Billing & Invoicing](#billing--invoicing)
9. [Reports & Analytics](#reports--analytics)
10. [Settings & Configuration](#settings--configuration)
11. [Advanced Features](#advanced-features)
12. [Troubleshooting](#troubleshooting)

## Getting Started

### First Login

1. **Navigate to your Yggdrasil instance**
   - URL: `https://your-domain.yggdrasil.app`
   - Use credentials provided by your organization

2. **Two-Factor Authentication (if enabled)**
   - Enter your authentication code from mobile app
   - Save browser if trusted device

3. **Welcome Tour**
   - Complete the guided tour (takes 2-3 minutes)
   - Review key features and navigation

### Dashboard Orientation

The Yggdrasil dashboard is organized into these main areas:

```bash
┌─────────────────────────────────────────────────────┐
│  Header: Navigation + User Menu + Notifications          │
├─────────────────────────────────────────────────────┤
│  Sidebar Navigation                                      │
│  ├─ 🏢 Dashboard       - Overview & metrics             │
│  ├─ 👥 Users           - User management              │
│  ├─ 🏢 Customers       - Client management           │
│  ├─ 🎫 Tickets          - Ticket management           │
│  ├─ 💻 Systems          - Managed systems             │
│  ├─ 💰 Billing          - Invoicing & payments       │
│  ├─ 📊 Reports         - Analytics & reports        │
│  └─ ⚙️ Settings         - System configuration        │
├─────────────────────────────────────────────────────┤
│  Main Content Area                                     │
└─────────────────────────────────────────────────────┘
```

## Understanding Your Role

### Business Administrator

**Primary Responsibilities:**

- Client relationship management
- User permission management
- Billing oversight and approval
- Report analysis and business insights
- Service configuration and pricing

**Key Features:**

- Customer onboarding
- Service plan management
- Invoice review and approval
- Business performance metrics
- Contract management

### System Administrator

**Primary Responsibilities:**

- System configuration and maintenance
- User technical support
- Integration management
- Security settings
- Performance monitoring

**Key Features:**

- System health monitoring
- Integration configuration
- Security policy enforcement
- User technical management
- Backup and recovery

## Dashboard Overview

### Quick Stats Cards

```markdown
| Metric              | Description                | Business Admin | System Admin |
| ------------------- | -------------------------- | -------------- | ------------ |
| **Active Users**    | Current logged-in users    | 156            | 142          |
| **Open Tickets**    | Unresolved support tickets | 23             | 45           |
| **Monthly Revenue** | Current month billing      | $45,280        | N/A          |
| **System Health**   | Platform uptime status     | 99.8%          | 99.9%        |
| **New Customers**   | Customers added this month | 12             | N/A          |
| **Scheduled Tasks** | Automation tasks           | 8              | 15           |
```

### Navigation Sections

#### **Dashboard Tab**

- **Real-time metrics**: Users online, tickets created, revenue today
- **Performance charts**: Ticket trends, revenue growth, system usage
- **Quick actions**: Create user, add customer, new ticket
- **Recent activity**: Last 10 system events

#### **Quick Access Panel**

- **Recently accessed**: Jump to recently viewed items
- **Saved searches**: Quick access to common reports
- **Bookmarked items**: Frequently accessed customers/tickets

## User Management

### Creating Users

#### Step-by-Step Process

1. **Navigate to Users** → **Add User**
2. **Fill User Information**:

   ```bash
   Required Fields:
   - Email address (unique)
   - Full name
   - Role selection
   - Customer assignment (if applicable)

   Optional Fields:
   - Phone number
   - Department
   - Manager
   - Time zone
   ```

3. **Configure Permissions** (based on role):

   ```markdown
   Available Roles:

   - **Super Admin**: Full system access
   - **Tenant Admin**: Customer management + billing
   - **Manager**: Team management + reports
   - **Technician**: Ticket management + assigned systems
   - **Client User**: Own tickets + basic info
   - **Read-only**: View access only
   ```

4. **Set Initial Password**:
   - Temporary password generated
   - Force change on first login
   - Send welcome email automatically

5. **Review and Create**:
   - Review all entered information
   - Click "Create User"
   - User receives onboarding email

#### Bulk User Import

**For adding multiple users:**

1. **Download Template**: Go to Users → Import → Download CSV Template
2. **Prepare CSV File**:

   ```csv
   email,name,role,customer_id,department,phone
   john.doe@company.com,John Doe,Technician,123,Support,+1-555-0123
   jane.smith@company.com,Jane Smith,Manager,123,Support,+1-555-0124
   ```

3. **Import File**:
   - Upload CSV file
   - Map columns to fields
   - Preview before final import
   - Review import results

4. **Post-Import**:
   - Send welcome emails to new users
   - Verify successful account creation
   - Handle any import errors

### Managing User Permissions

#### Role-Based Access Control (RBAC)

**Super Admin:**

- All system features
- User management
- Customer management
- Billing and invoicing
- System configuration
- Reports and analytics
- API access management

**Tenant Admin:**

- User management (own customers)
- Customer management
- Billing and invoicing
- Reports (own customers)
- Service plan management

**Manager:**

- User management (own team)
- Ticket management (team)
- System access (assigned systems)
- Team performance reports
- Work log management

**Technician:**

- Ticket management (assigned)
- System access (assigned systems)
- Time tracking
- Work log entries
- Basic customer information

#### Permission Granularity

```markdown
Fine-Grained Permissions:
├─ Users
│ ├─ Create users
│ ├─ Edit own profile
│ ├─ View team users
│ └─ Manage team users
├─ Tickets
│ ├─ Create tickets
│ ├─ Edit own tickets
│ ├─ Edit team tickets
│ ├─ Close tickets
│ └─ Delete tickets
├─ Systems
│ ├─ View assigned systems
│ ├─ Edit system info
│ ├─ Install agents
│ └─ Manage monitoring
└─ Reports
├─ View own performance
├─ View team performance
├─ View customer reports
└─ Export data
```

### User Lifecycle Management

#### Onboarding Workflow

1. **Account Creation**: As shown above
2. **Welcome Email**: Automatic with login credentials
3. **Role Training**: Role-specific onboarding materials
4. **Mentor Assignment**: Assign to experienced team member
5. **Progress Tracking**: Monitor onboarding completion
6. **Performance Review**: Initial 30-day review

#### Deactivation Workflow

1. **Ticket Assignment**: Reassign open tickets
2. **System Access**: Remove system permissions
3. **Data Handoff**: Transfer reports and data
4. **Customer Notification**: Inform affected customers
5. **Account Archive**: Deactivate and archive user data

## Customer Management

### Adding New Customers

#### Customer Creation Process

1. **Navigate to Customers** → **Add Customer**
2. **Basic Information**:

   ```bash
   Required:
   - Company name
   - Primary contact email
   - Service address

   Optional:
   - Phone number
   - Website
   - Industry
   - Company size
   ```

3. **Service Configuration**:

   ```markdown
   Service Plan Selection:

   - Basic: Core ticketing + basic monitoring
   - Professional: All basic + priority support + SLA
   - Enterprise: All professional + dedicated support + API access

   Billing Cycle:

   - Monthly
   - Quarterly
   - Annual
   - Custom
   ```

4. **Technical Setup**:
   - Assign primary contact as client user
   - Create initial administrator account
   - Configure service level agreement (SLA)
   - Set notification preferences

5. **Welcome Process**:
   - Send welcome email with credentials
   - Schedule onboarding call
   - Create initial ticket for setup
   - Assign account manager

#### Customer Information Management

#### Company Profile Management

```markdown
Customer Profile Sections:
├─ Company Information
│ ├─ Legal name
│ ├─ DBA/Trading name
│ ├─ Tax ID / EIN
│ ├─ Industry classification
│ └─ Company size (employees/revenue)
├─ Contact Information
│ ├─ Primary contact
│ ├─ Billing contact
│ ├─ Technical contact
│ └─ Emergency contacts
├─ Service Information
│ ├─ Current service plan
│ ├─ Start date
│ ├─ Contract term
│ ├─ SLA requirements
│ └─ Custom service add-ons
└─ Billing Information
├─ Billing method
├─ Payment terms
├─ Credit limit
└─ Invoice preferences
```

#### Service Plan Management

**Available Service Plans:**

1. **Basic Plan**:
   - Ticket management: 25 tickets/month
   - System monitoring: Up to 10 systems
   - Response time: 48 hours
   - Support: Email only
   - Price: $99/month

2. **Professional Plan**:
   - Ticket management: 100 tickets/month
   - System monitoring: Up to 50 systems
   - Response time: 24 hours
   - Support: Email + phone
   - API access: Basic read/write
   - Reports: Standard analytics
   - Price: $299/month

3. **Enterprise Plan**:
   - Ticket management: Unlimited
   - System monitoring: Unlimited
   - Response time: 4 hours (business hours)
   - Support: 24/7 dedicated
   - API access: Full API + webhook support
   - Reports: Advanced analytics + custom reports
   - Dedicated account manager
   - Custom SLA available
   - Price: Custom quote

#### Customer Communication

**Communication Preferences:**

- Email notifications: Ticket updates, system alerts, billing
- SMS notifications: Critical tickets, system outages
- In-app notifications: Real-time updates
- Monthly reports: Performance summary, upcoming maintenance

**Customer Portal Access:**

- Self-service ticket creation
- View own tickets and systems
- Access knowledge base
- Download invoices and reports
- Update contact information

## Ticket Management

### Creating and Managing Tickets

#### Ticket Creation Workflow

1. **New Ticket Creation**:

   ```bash
   Ticket Information:
   - Customer (required)
   - Title (required, max 200 chars)
   - Description (required)
   - Priority (1=Highest to 5=Lowest)
   - Category (Hardware, Software, Network, Other)
   - Impact (Low, Medium, High, Critical)
   - Assigned technician (optional)
   - Estimated hours (optional)
   - Due date (optional for SLA)
   - Attachments (max 10MB each)
   ```

2. **Automatic Assignment Rules**:

   ```markdown
   Assignment Logic:

   - By category: Network tickets to Network team
   - By customer: Preferred technician
   - By workload: Least tickets assigned
   - By skill: Match technician certifications
   - By availability: Currently online users
   - Escalation: Unassigned > 2 hours → Auto-escalate
   ```

#### Ticket Lifecycle Management

**Ticket States:**

1. **Open**: Created but not yet assigned
2. **Assigned**: Assigned to technician, work in progress
3. **In Progress**: Technician actively working on issue
4. **Pending**: Waiting for customer action or parts
5. **Resolved**: Issue fixed, awaiting customer confirmation
6. **Closed**: Customer confirmed resolution, ticket complete
7. **Cancelled**: Duplicate or invalid ticket

#### Work Time Tracking

**Time Entry Methods:**

- **Manual Entry**: Technician enters time spent
- **Timer Integration**: Built-in timer for active work
- **Mobile App**: Time tracking from mobile devices
- **Automatic Logging**: System-based activity tracking

**Time Categories:**

- Travel time (to/from customer site)
- On-site work time
- Remote work time
- Research and diagnosis
- Communication time

### SLA Management

#### Service Level Agreements

**Response Time SLAs:**

```markdown
Priority-Based SLAs:
├─ Priority 1 (Critical): 4 hours response, 24 hours resolution
├─ Priority 2 (High): 8 hours response, 48 hours resolution
├─ Priority 3 (Medium): 24 hours response, 72 hours resolution
├─ Priority 4 (Low): 48 hours response, 1 week resolution
└─ Priority 5 (Planning): 1 week response, 2 weeks resolution
```

**SLA Monitoring:**

- Real-time SLA tracking
- Automated alerts for SLA breaches
- SLA compliance reports
- Customer SLA dashboard
- Performance metrics by priority

### Ticket Automation

#### Automated Workflows

1. **Ticket Routing**:
   - Category-based routing
   - Customer preference routing
   - Priority-based escalation
   - Workload-based distribution

2. **Status Updates**:
   - Auto-assign based on rules
   - Auto-escalate overdue tickets
   - Customer notifications on status changes
   - Technician reminders for follow-up

3. **Resolution Templates**:
   - Common issue solutions
   - Standard response templates
   - Knowledge base integration
   - Auto-close resolved tickets

## System & Agent Management

### System Registration

#### Adding New Systems

1. **Navigate to Systems** → **Add System**
2. **System Information**:

   ```bash
   Required:
   - Hostname (unique)
   - IP Address
   - Operating System
   - Customer assignment
   - Location/Physical address

   Technical:
   - CPU information
   - Memory specifications
   - Storage details
   - Network configuration
   - Software inventory
   ```

3. **Agent Installation**:
   - Download agent software
   - Generate installation token
   - Provide installation instructions
   - Monitor installation progress
   - Verify connectivity to control plane

#### System Monitoring

**Metrics Collection:**

```markdown
Monitored Metrics:
├─ Performance Metrics
│ ├─ CPU usage (historical + real-time)
│ ├─ Memory usage
│ ├─ Disk usage and I/O
│ ├─ Network bandwidth
│ └─ Application performance
├─ Availability Metrics
│ ├─ Uptime monitoring
│ ├─ Service response times
│ ├─ Error rates
│ └─ Service availability
└─ Security Metrics
├─ Failed login attempts
├─ Unusual activity patterns
├─ Firewall block events
└─ Vulnerability scan results
```

#### Alert Configuration

**Alert Types:**

- **System Down**: Immediate notification
- **High Resource Usage**: Warning threshold exceeded
- **Performance Degradation**: Response times increased
- **Security Events**: Failed attempts, suspicious activity
- **Maintenance Windows**: Scheduled downtime notifications

**Alert Channels:**

- Email notifications to assigned technicians
- SMS for critical alerts
- In-app notifications for active users
- Webhook integrations to monitoring systems

### Agent Management

#### Agent Deployment

**Supported Platforms:**

- Windows Server (2016, 2019, 2022)
- Windows Desktop (10, 11)
- Linux (Ubuntu 18.04+, CentOS 7+, RHEL 7+)
- macOS (10.15+)

**Installation Methods:**

1. **Manual Installation**: Download and run installer
2. **Remote Deployment**: Push installation via management console
3. **Scripted Deployment**: Silent installation for mass deployment
4. **Container Deployment**: Docker/VM deployment options

#### Agent Configuration

**Configuration Options:**

- **Monitoring Frequency**: 1 minute to 1 hour intervals
- **Data Retention**: Local storage duration
- **Network Settings**: Proxy configuration for corporate environments
- **Security Settings**: Certificate validation, encryption keys
- **Performance Settings**: Resource limits, monitoring scope

## Billing & Invoicing

### Invoice Generation

#### Monthly Billing Process

1. **Usage Calculation** (end of month):

   ```bash
   Billable Items:
   ├─ User licenses (per seat)
   ├─ System monitoring (per system)
   ├─ Ticket volume (above included amount)
   ├─ Service add-ons (backup, etc.)
   ├─ SLA premium (if applicable)
   └─ Professional services (if applicable)
   ```

2. **Invoice Generation**:
   - Auto-generate draft invoices
   - Review for accuracy
   - Apply discounts or credits
   - Generate final invoice
   - Send to customer via email
   - Record payment due date

3. **Invoice Templates:**
   - Company branding
   - Itemized charges
   - Payment terms
   - Tax calculations
   - Currency formatting

#### Payment Processing

**Payment Methods:**

- Credit Card (automatic recurring)
- Bank Transfer (ACH/Wire)
- PayPal/Stripe integration
- Purchase orders (for enterprise)
- Custom payment terms

**Payment Workflow:**

1. **Invoice Sent**: Customer receives invoice
2. **Payment Due**: Based on terms (Net 15, Net 30)
3. **Reminders**: Automated reminders at 3, 7, 14 days overdue
4. **Payment Received**: Automatic notification and update
5. **Late Fees**: Applied based on terms
6. **Collections**: Escalation for significantly overdue accounts

### Financial Reporting

**Revenue Reports:**

- Monthly recurring revenue (MRR)
- Annual contract value (ACV)
- Customer lifetime value (CLV)
- Churn analysis and retention
- Revenue by service plan
- Geographic revenue distribution

**Cost Analysis:**

- Service delivery costs
- Customer acquisition costs
- Support overhead analysis
- Profitability by customer
- Cost per ticket metric

## Reports & Analytics

### Dashboard Analytics

#### Key Performance Indicators

**Business Metrics:**

```markdown
Business KPIs:
├─ Customer Metrics
│ ├─ Total customers
│ ├─ New customers (monthly)
│ ├─ Customer churn rate
│ ├─ Customer satisfaction scores
│ └─ Average contract value
├─ Service Metrics
│ ├─ Tickets created/resolved
│ ├─ Average response time
│ ├─ Average resolution time
│ ├─ First contact resolution rate
│ └─ SLA compliance percentage
├─ Financial Metrics
│ ├─ Monthly recurring revenue
│ ├─ Revenue growth rate
│ ├─ Customer acquisition cost
│ ├─ Average revenue per customer
│ └─ Profit margins by service
└─ Operational Metrics
├─ System uptime percentage
├─ Agent deployment status
├─ User activity levels
├─ Support team productivity
└─ Resource utilization
```

#### Custom Reports

**Report Builder Features:**

- Drag-and-drop report designer
- Pre-built report templates
- Custom metric calculations
- Data filtering and segmentation
- Scheduled report generation
- Multiple export formats (PDF, Excel, CSV)

**Popular Report Templates:**

- Monthly performance summary
- Customer health scorecard
- Team productivity report
- Revenue analysis dashboard
- System utilization report
- SLA compliance report
- Ticket volume analysis

### Data Visualization

**Chart Types:**

- Line charts: Trends over time
- Bar charts: Comparative analysis
- Pie charts: Distribution analysis
- Heat maps: Geographic distribution
- Gauges: Real-time metrics
- Tables: Detailed data breakdown

**Interactive Features:**

- Drill-down capabilities
- Real-time data updates
- Filter by date range, customer, team
- Export chart as image
- Share reports via link
- Schedule automatic delivery

## Settings & Configuration

### System Configuration

#### General Settings

```markdown
System Configuration Sections:
├─ Company Settings
│ ├─ Company name and logo
│ ├─ Contact information
│ ├─ Time zone and locale
│ └─ Business hours
├─ Security Settings
│ ├─ Password policies
│ ├─ Session timeout settings
│ ├─ Two-factor authentication
│ ├─ IP restrictions
│ └─ Audit logging configuration
├─ Email Settings
│ ├─ SMTP server configuration
│ ├─ Email templates
│ ├─ Notification preferences
│ └─ Bounce handling
└─ Integration Settings
├─ API key management
├─ Webhook configuration
├─ Single sign-on (SSO)
└─ Third-party integrations
```

#### Security Configuration

**Password Policies:**

- Minimum length: 8+ characters
- Complexity requirements: Upper, lower, number, symbol
- Expiration: 90 days
- History: Prevent reuse of last 5 passwords
- Lockout: 5 failed attempts, 15 minute lockout

**Session Management:**

- Default timeout: 8 hours
- Remember me: 30 days
- Concurrent sessions: Maximum 3 per user
- IP restrictions: Optional whitelist/blacklist
- Device management: Recognize trusted devices

### Integration Management

#### API Access

**API Key Management:**

- Generate API keys for each integration
- Set specific permissions per key
- Rate limiting per key
- Usage monitoring and quotas
- Key rotation and expiration

**Webhook Configuration:**

- Event types: Tickets, users, systems, billing
- Endpoint URL configuration
- Signature verification setup
- Retry policies and error handling
- Event payload customization

#### Third-Party Integrations

**Common Integrations:**

- **Accounting**: QuickBooks, Xero
- **Monitoring**: Datadog, New Relic, PagerDuty
- **Communication**: Slack, Microsoft Teams
- **CRM**: Salesforce, HubSpot
- **Backup**: AWS S3, Backblaze

## Advanced Features

### Automation Workflows

#### Business Process Automation

**Ticket Automation:**

- Auto-assignment based on skill matching
- Escalation rules for overdue tickets
- Customer notification workflows
- Resolution confirmation automation

**Customer Lifecycle:**

- Automated onboarding sequences
- Renewal reminder workflows
- Churn prevention alerts
- Satisfaction survey automation

#### Custom Scripting

**Available Scripting:**

- Custom validation rules
- Automated report generation
- Data transformation scripts
- Integration with external APIs
- Custom notification logic

### Multi-Tenant Management

**Tenant Isolation:**

- Complete data separation
- Custom branding per tenant
- Independent user management
- Separate billing and invoicing
- Tenant-specific integrations

**Global Management:**

- Cross-tenant reporting
- Global user management
- Service plan updates
- System-wide announcements
- Tenant health monitoring

## Troubleshooting

### Common Issues

#### User Management Issues

**Problem**: User cannot log in
**Solutions**:

1. Verify correct username/password
2. Check if account is locked (too many failed attempts)
3. Verify account is active (not deactivated)
4. Check two-factor authentication setup
5. Clear browser cache and cookies

**Problem**: Missing permissions error
**Solutions**:

1. Verify user role and assigned permissions
2. Check if permissions apply to specific customer
3. Verify session hasn't expired
4. Check if user needs additional training
5. Contact system administrator if issue persists

#### System Monitoring Issues

**Problem**: Agent not reporting data
**Solutions**:

1. Check network connectivity between agent and control plane
2. Verify agent installation completed successfully
3. Check firewall rules allow outbound connections
4. Verify agent configuration settings
5. Restart agent service

**Problem**: False alerts or notifications
**Solutions**:

1. Adjust alert thresholds
2. Configure maintenance windows to suppress alerts
3. Review and fine-tune alert rules
4. Add alert suppression rules for known issues
5. Update monitoring configurations

#### Billing Issues

**Problem**: Invoice amounts incorrect
**Solutions**:

1. Review customer service plan configuration
2. Check for manual adjustments or credits
3. Verify billing period calculation
4. Review add-on services pricing
5. Check tax configuration for customer location

**Problem**: Payment processing failures
**Solutions**:

1. Verify payment method is valid and current
2. Check payment processor status
3. Verify billing information accuracy
4. Contact customer for updated payment method
5. Review transaction error logs

### Performance Optimization

#### System Performance

**Dashboard Optimization**:

- Use data filtering to reduce load times
- Schedule large report generation during off-peak hours
- Implement data caching for frequently accessed reports
- Optimize database queries for dashboard metrics
- Use content delivery networks for static assets

**User Experience**:

- Implement progressive loading for large datasets
- Optimize mobile interface for administrators
- Provide keyboard shortcuts for common actions
- Implement search with auto-complete
- Use browser local storage for user preferences

#### Database Maintenance

**Regular Maintenance Tasks**:

- Index optimization for frequently queried fields
- Archive historical data to improve query performance
- Regular backup verification and testing
- Monitor and optimize slow queries
- Update database statistics regularly

### Getting Help

#### Support Channels

**Documentation Resources:**

- [User Guide](./USER_GUIDE.md) - End-user documentation
- [API Documentation](../api/QUICK_START.md) - Integration guides
- [Developer Documentation](../DEVELOPMENT_SETUP.md) - Technical documentation
- [Knowledge Base](https://support.yggdrasil.app/kb) - Self-service articles

**Community Support:**

- User forums for peer support
- Best practices sharing
- Feature request submission
- Bug reporting and tracking
- User group meetings and webinars

**Technical Support:**

- Support ticket system
- Phone support (enterprise plans)
- Email support with SLA guarantees
- Remote assistance for critical issues
- Emergency contact information

---

This administrator guide provides comprehensive coverage of both business and technical administration tasks. For additional assistance, contact your organization's Yggdrasil administrator or open a support ticket through the platform.
