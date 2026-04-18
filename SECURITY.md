# Security Policy

## Reporting a Vulnerability

If you believe you have found a security vulnerability in llmstatus.io,
please report it responsibly.

**Do not** open a public GitHub issue for security vulnerabilities.

**Do** email `security@llmstatus.io` with:

- A description of the vulnerability
- Steps to reproduce
- Potential impact
- Any suggested mitigation

We commit to:

- Acknowledging your report within 48 hours
- Providing a preliminary assessment within 7 days
- Keeping you informed of remediation progress
- Giving you credit (if you wish) once the issue is resolved

We ask that you:

- Give us a reasonable time (typically 30 days) to resolve the issue
  before public disclosure
- Make a good faith effort to avoid data destruction, privacy violations,
  or service interruption during your research
- Not exploit the vulnerability beyond what is necessary to demonstrate it

---

## Scope

The following are **in scope** for security reports:

- The llmstatus.io website and API
- The probe infrastructure and ingestion pipeline
- Authentication, authorization, and session management
- Data exposure or injection vulnerabilities
- Denial-of-service vulnerabilities in our systems
- Dependency vulnerabilities we haven't patched

The following are **out of scope**:

- Vulnerabilities in third-party AI provider APIs we monitor (report to them)
- Issues requiring physical access to our infrastructure
- Issues requiring social engineering of our team
- Rate-limiting bypass that doesn't affect data integrity
- Cosmetic issues or typos

---

## Not a Security Issue

The following are NOT security issues, even if surprising:

- **We publish incident data about other companies** — this is our
  stated methodology, not a leak
- **Our source code is public** — this is deliberate, see
  `METHODOLOGY.md` §13
- **We publish historical data permanently** — this is a feature, not
  a bug

If you're unsure whether something is a security issue, email us and ask.

---

## Responsible Disclosure Recognition

We maintain a list of researchers who have responsibly disclosed issues
to us at https://llmstatus.io/security/acknowledgments (with permission).

---

## Preferred Languages

We accept reports in English and Chinese.
