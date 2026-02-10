# Yggdrasil User Guide

## Welcome to Yggdrasil

Yggdrasil is a modular web application platform designed for scalability, security, and rapid development. This guide will walk you through everything you need to know to use Yggdrasil effectively.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Getting Started](#getting-started)
3. [Basic Usage](#basic-usage)
4. [Managing Users and Permissions](#managing-users-and-permissions)
5. [Working with Data](#working-with-data)
6. [Plugins and Extensions](#plugins-and-extensions)
7. [Monitoring and Troubleshooting](#monitoring-and-troubleshooting)
8. [Advanced Features](#advanced-features)

---

## Quick Start

### Prerequisites

- Node.js 18+ and npm
- PostgreSQL 14+ (for production)
- Git

### Installation Steps

1. **Clone the repository**

   ```bash
   git clone https://github.com/intrik8-labs/yggdrasil.git
   cd yggdrasil
   ```

2. **Install dependencies**

   ```bash
   npm install
   ```

3. **Set up environment**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Initialize database**

   ```bash
   npm run db:migrate
   npm run db:seed
   ```

5. **Start the application**

   ```bash
   npm run dev
   ```

6. **Access the application**
   - Open http://localhost:3000
   - Default admin credentials: `admin / admin123`

That's it! You now have Yggdrasil running locally.

---

## Getting Started

### Understanding the Interface

When you first log in to Yggdrasil, you'll see the main dashboard with:

- **Navigation Menu**: Access all major sections
- **Dashboard Overview**: System status and quick actions
- **User Profile**: Account settings and preferences
- **Help Center**: Documentation and support resources

### Basic Navigation

1. **Main Menu** (left sidebar)
   - Dashboard: Home screen with overview
   - Users: User management
   - Content: Data and content management
   - Plugins: Extension management
   - Settings: System configuration
   - Admin: Administrative tools

2. **Quick Actions** (top bar)
   - Create new content/items
   - Search functionality
   - Notifications
   - User menu

### Your First Steps

1. **Update Your Profile**
   - Click your name in the top-right corner
   - Select "Profile Settings"
   - Update your name, email, and password

2. **Explore the Dashboard**
   - Review the system overview
   - Check available quick actions
   - Familiarize yourself with the layout

3. **Create Your First Item**
   - Navigate to "Content" in the menu
   - Click "Create New"
   - Fill in the required fields
   - Save your changes

---

## Basic Usage

### Creating and Managing Content

#### Creating New Content

1. **Navigate to Content Section**

   ```
   Main Menu → Content → Create New
   ```

2. **Fill in Content Details**
   - **Title**: Descriptive name for your content
   - **Type**: Select appropriate content type
   - **Description**: Detailed information
   - **Tags**: Organize with keywords

3. **Configure Settings**
   - **Visibility**: Public, private, or restricted
   - **Permissions**: Who can view/edit
   - **Publishing**: Schedule or publish immediately

#### Editing Content

1. **Find your content** in the Content list
2. **Click the edit icon** (pencil) next to the item
3. **Make your changes**
4. **Save** to update or **Cancel** to discard changes

#### Deleting Content

1. **Select content** from the list
2. **Click the delete icon** (trash)
3. **Confirm** the deletion action

### Search and Filter

#### Basic Search

- **Quick Search**: Use the search bar in the top navigation
- **Results**: Shows matching content across all areas

#### Advanced Filtering

1. **Go to Content section**
2. **Use filter options**:
   - Date range
   - Content type
   - Tags
   - Author/Creator
   - Status (published, draft, archived)

### Working with Lists and Tables

#### Sorting

- **Click column headers** to sort ascending/descending
- **Multiple sort**: Hold Shift + click for secondary sorting

#### Bulk Operations

1. **Select items** using checkboxes
2. **Choose action** from the bulk actions menu:
   - Delete selected
   - Change permissions
   - Export data
   - Archive items

---

## Managing Users and Permissions

### User Roles and Permissions

Yggdrasil uses a role-based access control system:

#### Default Roles

1. **Administrator**
   - Full system access
   - User management
   - System configuration
   - Plugin management

2. **Editor**
   - Create and edit content
   - Moderate user-generated content
   - Manage categories and tags

3. **Author**
   - Create and own content
   - Edit own content
   - View published content

4. **Viewer** (Read-only)
   - View published content
   - Basic search and browsing

#### Custom Roles

Administrators can create custom roles with specific permissions:

1. **Navigate to Settings → Roles**
2. **Click "Create Role"**
3. **Select permissions** from available categories
4. **Save the role**

### User Management

#### Adding New Users

1. **Go to Users section** in the main menu
2. **Click "Add User"**
3. **Fill in user information**:
   - Name and email
   - Username
   - Initial password
   - Role assignment
4. **Send invitation** (optional)

#### Editing User Accounts

1. **Find the user** in the Users list
2. **Click "Edit"** next to their name
3. **Update information** as needed
4. **Save changes**

#### Managing User Permissions

1. **Select user** from Users list
2. **Click "Permissions"** tab
3. **Modify role** or assign custom permissions
4. **Apply changes**

### Security Best Practices

#### Password Policies

- **Minimum length**: 8 characters
- **Complexity**: Mix of letters, numbers, symbols
- **Expiration**: Regular password changes (configurable)

#### Session Management

- **Auto-logout**: After inactivity period
- **Session timeout**: Configurable time limit
- **Concurrent sessions**: Limit per user

#### Access Control

- **Two-factor authentication**: Optional but recommended
- **IP restrictions**: Limit access by IP range
- **Audit logging**: Track user actions

---

## Working with Data

### Data Import and Export

#### Importing Data

1. **Prepare your data file** (CSV, JSON, or XML format)
2. **Navigate to Settings → Data Management**
3. **Click "Import Data"**
4. **Select file** and configure import options:
   - Data type mapping
   - Field matching
   - Update/merge options
5. **Preview and confirm** import

#### Exporting Data

1. **Go to Content section**
2. **Filter and select** items to export
3. **Click "Export"**
4. **Choose format** (CSV, JSON, XML, PDF)
5. **Select fields** to include
6. **Download the file**

### Data Backup and Recovery

#### Automated Backups

Yggdrasil supports automated backups:

1. **Navigate to Settings → Backup**
2. **Configure backup schedule**:
   - Daily, weekly, or monthly
   - Retention period
   - Storage location
3. **Enable automatic backups**

#### Manual Backups

1. **Go to Settings → Backup**
2. **Click "Create Backup Now"**
3. **Choose backup type**:
   - Full backup (database + files)
   - Database only
   - Files only
4. **Download or store** backup file

#### Restoring from Backup

⚠️ **Warning**: Restoration will overwrite current data

1. **Go to Settings → Backup**
2. **Click "Restore from Backup"**
3. **Upload backup file** or select from stored backups
4. **Confirm restoration**
5. **Wait for process completion**

### Data Validation and Quality

#### Data Validation Rules

Yggdrasil automatically validates data based on:

- **Field types**: Text, numbers, dates, etc.
- **Required fields**: Mandatory information
- **Format validation**: Email, URLs, phone numbers
- **Custom rules**: Business-specific validation

#### Data Cleaning Tools

1. **Navigate to Settings → Data Management**
2. **Use cleaning tools**:
   - Remove duplicates
   - Standardize formats
   - Fix invalid entries
   - Archive old data

---

## Plugins and Extensions

### Understanding the Plugin System

Yggdrasil's plugin architecture allows extending functionality without modifying core code:

#### Plugin Types

1. **Content Plugins**: Add new content types and fields
2. **Integration Plugins**: Connect with external services
3. **UI Plugins**: Modify interface and user experience
4. **Security Plugins**: Add authentication and authorization methods
5. **Analytics Plugins**: Track usage and generate reports

### Installing Plugins

#### From Plugin Store

1. **Navigate to Plugins section**
2. **Browse Plugin Store**
3. **Find desired plugin**
4. **Click "Install"**
5. **Configure plugin settings**
6. **Enable the plugin**

#### Manual Installation

1. **Download plugin files** (.zip or git repository)
2. **Navigate to Plugins → Install Plugin**
3. **Upload plugin file** or provide repository URL
4. **Follow installation wizard**
5. **Configure and enable**

### Managing Installed Plugins

#### Plugin Configuration

1. **Go to Plugins section**
2. **Select plugin** from the list
3. **Access settings panel**
4. **Configure options**:
   - General settings
   - Permissions
   - Integration options
   - Advanced features

#### Enabling/Disabling Plugins

1. **Find plugin** in Plugins list
2. **Toggle enable/disable switch**
3. **Confirm action** (some plugins require restart)

#### Updating Plugins

1. **Check for updates** in Plugins section
2. **Select plugins** to update
3. **Click "Update Selected"**
4. **Review changes** and confirm
5. **Restart if required**

#### Removing Plugins

1. **Select plugin** to remove
2. **Click "Uninstall"**
3. **Choose data handling**:
   - Keep plugin data
   - Remove all plugin data
4. **Confirm removal**

### Developing Custom Plugins

For developers interested in creating custom plugins, see the [Development Guide](../architecture/README.md) and [Plugin Architecture](../architecture/backend/plugin-architecture.md) documentation.

---

## Monitoring and Troubleshooting

### System Monitoring

#### Dashboard Overview

The main dashboard provides real-time system information:

- **System Health**: Overall status indicator
- **Performance Metrics**: Response times, CPU, memory usage
- **Active Users**: Current logged-in users
- **Recent Activities**: Latest system events
- **Storage Usage**: Database and file storage statistics

#### Detailed Monitoring

1. **Navigate to Admin → Monitoring**
2. **View detailed metrics**:
   - Performance graphs
   - Error rates
   - Database performance
   - Cache statistics
   - Network usage

### Logs and Debugging

#### Accessing Logs

1. **System Logs**: Admin → Logs → System
2. **Error Logs**: Admin → Logs → Errors
3. **Audit Logs**: Admin → Logs → Audit
4. **Performance Logs**: Admin → Logs → Performance

#### Log Filtering and Search

- **Date range**: Filter by time period
- **Log level**: Error, warning, info, debug
- **Component**: System modules and plugins
- **User**: Filter by specific user actions
- **Search**: Text search within log entries

#### Common Issues and Solutions

##### Login Problems

**Issue**: Can't log in with correct credentials
**Solutions**:

1. Check password spelling and case
2. Verify account is not locked
3. Clear browser cache and cookies
4. Try password reset if available

##### Performance Issues

**Issue**: Slow page loading or response times
**Solutions**:

1. Check system monitoring for bottlenecks
2. Clear cache: Admin → Maintenance → Clear Cache
3. Check database connections
4. Review recent plugin installations

##### Plugin Errors

**Issue**: Plugin not working or causing errors
**Solutions**:

1. Disable problematic plugin
2. Check plugin compatibility
3. Review error logs for specific issues
4. Update to latest version

##### Data Import Failures

**Issue**: Data import process fails
**Solutions**:

1. Verify file format is correct
2. Check field mapping configuration
3. Ensure data meets validation rules
4. Split large files into smaller chunks

### Maintenance Tasks

#### Scheduled Maintenance

1. **Navigate to Admin → Maintenance**
2. **Schedule regular tasks**:
   - Database optimization
   - Cache clearing
   - Log rotation
   - Backup verification

#### Manual Maintenance

1. **Clear System Cache**

   ```
   Admin → Maintenance → Clear Cache
   ```

2. **Optimize Database**

   ```
   Admin → Maintenance → Database Optimization
   ```

3. **Clean Up Old Data**
   ```
   Admin → Maintenance → Data Cleanup
   ```

---

## Advanced Features

### API Access

#### Getting API Credentials

1. **Navigate to Settings → API**
2. **Click "Generate API Key"**
3. **Set permissions and rate limits**
4. **Save and securely store** the API key

#### Making API Requests

```bash
# Example API request
curl -X GET "http://localhost:3000/api/v1/content" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json"
```

#### API Documentation

Complete API documentation is available at:

- **Interactive**: http://localhost:3000/api/docs
- **Static**: [API Quick Start](../api/QUICK_START.md)

### Workflow Automation

#### Creating Workflows

1. **Navigate to Settings → Workflows**
2. **Click "Create Workflow"**
3. **Define triggers**:
   - Content creation
   - User actions
   - Schedule-based
   - External events
4. **Add actions**:
   - Send notifications
   - Update data
   - Call external APIs
   - Generate reports
5. **Test and enable** the workflow

#### Common Workflow Examples

##### Content Review Process

1. **Trigger**: New content created
2. **Action**: Notify reviewers via email
3. **Condition**: Content type = "Article"
4. **Action**: Add to review queue
5. **Follow-up**: Notify author when approved

##### User Onboarding

1. **Trigger**: New user registration
2. **Action**: Send welcome email
3. **Action**: Assign default permissions
4. **Action**: Create user profile
5. **Follow-up**: Schedule follow-up after 7 days

### Custom Fields and Data Types

#### Creating Custom Fields

1. **Navigate to Settings → Content Types**
2. **Select content type** to modify
3. **Click "Add Field"**
4. **Configure field**:
   - Field name and type
   - Validation rules
   - Display options
   - Permissions
5. **Save and test** the field

#### Supported Field Types

- **Text**: Single line, multi-line, rich text
- **Numbers**: Integer, decimal, currency
- **Dates**: Date, datetime, time
- **Selection**: Dropdown, radio buttons, checkboxes
- **Media**: Images, files, videos
- **Relationships**: Links to other content
- **Calculated**: Computed values based on other fields

### Multi-tenant Configuration

#### Setting Up Tenants

For organizations managing multiple instances:

1. **Enable multi-tenant mode** in system settings
2. **Create tenant configuration**:
   - Unique identifier
   - Database schema
   - Custom domains
   - User isolation
3. **Configure tenant-specific settings**
4. **Test tenant isolation**

#### Tenant Management

- **Isolation**: Separate data and users per tenant
- **Customization**: Tenant-specific branding and features
- **Resource allocation**: Memory, storage, and user limits
- **Billing**: Individual or shared billing models

---

## Getting Help and Support

### Help Resources

1. **In-App Help**: Click the help icon (?) in the interface
2. **Documentation**: Browse comprehensive guides
3. **Community Forum**: Connect with other users
4. **Video Tutorials**: Step-by-step visual guides
5. **Knowledge Base**: Search FAQs and solutions

### Contact Support

#### Support Channels

- **Email**: support@yggdrasil-platform.com
- **Live Chat**: Available during business hours
- **Support Tickets**: Track and manage support requests
- **Phone**: Priority support for enterprise customers

#### Reporting Issues

When reporting problems, include:

1. **Description**: Detailed issue description
2. **Steps to reproduce**: Exact actions to replicate
3. **Expected vs actual**: What should happen vs what happened
4. **Environment**: Browser, operating system, version
5. **Error messages**: Complete error text or screenshots
6. **Logs**: Relevant system logs if available

### Community and Contributions

#### Join the Community

- **GitHub**: Contribute to development
- **Discord/Slack**: Real-time discussions
- **User Groups**: Local meetups and events
- **Blog**: Latest updates and best practices

#### Contributing to Documentation

Found something missing or unclear?

1. **Fork the documentation repository**
2. **Make improvements**
3. **Submit a pull request**
4. **Join the documentation team**

---

## Conclusion

This user guide covers the essential features and functionality of Yggdrasil. As you become more familiar with the platform, you'll discover additional capabilities and workflows that can be customized to your specific needs.

### Next Steps

1. **Explore advanced features** based on your use case
2. **Join the community** to learn from other users
3. **Consider custom plugins** for specialized requirements
4. **Stay updated** with new releases and features

### Additional Resources

- [Developer Documentation](../architecture/README.md)
- [API Reference](../api/QUICK_START.md)
- [Security Guide](../security/SECURITY.md)
- [Deployment Guide](../deployment/DEPLOYMENT.md)

Thank you for choosing Yggdrasil! We're excited to see what you'll build with this powerful platform.

### Next Steps

1. **Explore advanced features** based on your use case
2. **Join the community** to learn from other users
3. **Consider custom plugins** for specialized requirements
4. **Stay updated** with new releases and features

### Additional Resources

- [Developer Documentation](../architecture/README.md)
- [API Reference](../api/QUICK_START.md)
- [Security Guide](../security/SECURITY.md)
- [Deployment Guide](../deployment/DEPLOYMENT.md)

Thank you for choosing Yggdrasil! We're excited to see what you'll build with this powerful platform.
