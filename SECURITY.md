# Security Policy

## Supported Versions

We actively support the following versions with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please follow responsible disclosure practices:

### 🚨 For Security Issues

**DO NOT** open a public GitHub issue for security vulnerabilities.

Instead, please report security issues via:

1. **Email**: [security@rmfosho.me](mailto:security@rmfosho.me) (temp)
2. **GitHub Security Advisories**: [Create a security advisory](https://github.com/Native-Consulting-Services/chukfi-cms/security/advisories/new)

### 📋 Include in Your Report

Please include as much information as possible:

- **Description** of the vulnerability
- **Steps to reproduce** the issue
- **Potential impact** and attack scenarios
- **Affected versions** and components
- **Suggested fix** (if you have one)

### 🔒 Our Commitment

- **Initial Response**: Within 48 hours
- **Status Updates**: Every 7 days until resolved
- **Fix Timeline**: Critical issues within 7 days, others within 30 days
- **Credit**: We'll acknowledge your contribution (unless you prefer anonymity)

## Security Best Practices

### For Users

- Keep Chukfi CMS updated to the latest version
- Use strong, unique passwords for admin accounts
- Enable HTTPS in production environments
- Regularly backup your database
- Monitor access logs for suspicious activity
- Use environment variables for sensitive configuration

### For Developers

- Never commit sensitive information (passwords, API keys)
- Use parameterized queries to prevent SQL injection
- Validate all user input on both frontend and backend
- Follow JWT token best practices (short expiration, secure storage)
- Use HTTPS for all external API calls
- Regularly update dependencies

## Common Security Considerations

### Authentication

- JWT tokens have configurable expiration times
- Password hashing uses bcrypt with appropriate cost
- Session management follows security best practices

### Authorization

- Role-based access control (RBAC) system
- Permission checks on all protected routes
- Input validation on all API endpoints

### Database Security

- SQLite file permissions should be properly configured
- Database queries use parameterized statements
- No sensitive data logged in plain text

### Frontend Security

- XSS protection through proper input sanitization
- CSRF protection for state-changing operations
- Secure cookie configuration
- Content Security Policy (CSP) headers

## Third-Party Dependencies

We regularly monitor and update dependencies for security vulnerabilities:

- **Go modules**: Updated with `go mod tidy`
- **npm packages**: Monitored with `npm audit`
- **GitHub Security Advisories**: Automated monitoring enabled

## Vulnerability Response Process

1. **Receive** vulnerability report
2. **Verify** and assess impact
3. **Develop** and test fix
4. **Coordinate** disclosure timeline
5. **Release** security update
6. **Publish** security advisory
7. **Notify** users of the update

## Security Updates

Security updates are published as:

- **Patch releases** (e.g., v1.0.1) for minor security fixes
- **Minor releases** (e.g., v1.1.0) for significant security improvements
- **Major releases** (e.g., v2.0.0) for security-related breaking changes

Subscribe to [GitHub Releases](https://github.com/Native-Consulting-Services/chukfi-cms/releases) to be notified of security updates.

---

Thank you for helping keep Chukfi CMS secure! 🔒
