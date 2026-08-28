# Security Policy

Security is taken seriously in this project.

Please report suspected security vulnerabilities responsibly and avoid publicly disclosing vulnerabilities before they have been investigated and, where appropriate, resolved.

## Supported Versions

| Version | Supported |
|---|---|
| Latest release | Yes |
| Older releases | Best effort |
| Unsupported releases | No |

## Reporting a Vulnerability

**Do not report security vulnerabilities through public GitHub issues.**

The preferred reporting mechanism is GitHub Private Vulnerability Reporting, where enabled for this repository.

Navigate to **Repository → Security → Advisories → Report a vulnerability**.

If private vulnerability reporting is unavailable, contact the maintainer directly:

**Steve Owens**  
GitHub: @sowens81  
Email: stevejowens81@outlook.com

When reporting a vulnerability, please include a description, affected component or version, reproduction steps, potential impact, known mitigations, and safe proof-of-concept material where appropriate.

## Disclosure

Please allow reasonable time for the vulnerability to be investigated and addressed before public disclosure.

## Secrets

The repository must never contain passwords, API keys, access or refresh tokens, cloud credentials, private or signing keys, production certificates containing private keys, `.env` files containing secrets, or Terraform state containing sensitive information.

If a secret is accidentally committed, the affected credential should be considered compromised and immediately revoked or rotated.

## Dependencies

Dependencies should be kept reasonably current and reviewed for known security vulnerabilities.
