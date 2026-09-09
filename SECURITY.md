# Security Policy

YWD-APRS is currently pre-alpha software and should be treated accordingly.

## Supported versions

Until the project reaches stable releases, security fixes will target the current development line. Old pre-release checkpoints may not receive backports.

## Reporting a vulnerability

Please avoid publishing exploitable security details, credentials, private keys, or authentication material in a public issue.

For non-sensitive hardening suggestions and ordinary bugs, use the repository issue tracker.

For a vulnerability that would expose credentials, permit unauthorized administration, enable unintended RF transmission, corrupt operator data, or otherwise create a meaningful security/safety risk, use GitHub's private vulnerability reporting/security-advisory mechanism if it is enabled for the repository.

## Security-sensitive areas

Particular care is required around:

- web authentication and session handling
- remote administration
- update/download verification
- configuration files and secrets
- APRS-IS credentials
- filesystem permissions
- SQLite backup/restore paths
- browser-to-daemon command authorization
- RF transmit enable/disable state
- IGate Internet-to-RF policy
- parsing of untrusted KISS, AX.25, APRS, and network input

## Deployment guidance

During early development:

- prefer LAN-only access
- do not expose the web interface directly to the public Internet
- run the daemon with the minimum filesystem/device permissions required
- keep TX disabled when testing receive-only milestones
- back up configuration and database files before testing migrations or updater behavior

The project will add stronger deployment and authentication guidance as those features are implemented.
