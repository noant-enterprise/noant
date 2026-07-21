# Security Policy

## Supported Versions

| Version | Supported          |
|---------|--------------------|
| 2.0.x   | :white_check_mark: |
| < 2.0   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in NOANT Enterprise, please report it responsibly.

**DO NOT** open a public GitHub issue for security vulnerabilities.

### How to Report

1. Email **security@noant.example.com** with:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact assessment
   - Suggested fix (if any)

2. You should receive an acknowledgment within **48 hours**.

3. We will work with you to understand and address the issue before any public disclosure.

### What to Expect

- **Acknowledgment**: Within 48 hours of your report
- **Assessment**: Within 5 business days
- **Fix timeline**: Depends on severity
  - Critical: 24-48 hours
  - High: 1 week
  - Medium: 2 weeks
  - Low: Next release

### Scope

The following are in scope:
- Authentication/authorization bypass
- SQL injection
- Cross-site scripting (XSS)
- Remote code execution
- Data exposure
- CSRF attacks
- Rate limiting bypass
- Session hijacking

The following are out of scope:
- Denial of service (DoS) attacks
- Social engineering
- Physical attacks
- Issues in third-party dependencies (report to the dependency maintainer)

### Safe Harbor

We support responsible disclosure and will not take legal action against researchers who:
- Make a good faith effort to avoid privacy violations
- Do not exploit vulnerabilities beyond what is necessary to confirm their existence
- Do not access or modify other users' data
- Report promptly after discovery

### Bug Bounty

Currently, there is no formal bug bounty program. However, we recognize security researchers who help improve NOANT's security with:
- Public acknowledgment (with permission)
- Priority support for any issues encountered
