# Contributing

Thank you for considering contributing to this project.

Contributions including bug fixes, documentation improvements, new features, tests, and other improvements are welcome.

## Before Contributing

Before starting significant work:

1. Check existing issues and pull requests to determine whether the work is already being discussed or implemented.
2. For significant features or architectural changes, open an issue before implementation so the proposed approach can be discussed.
3. Keep changes focused on a single concern where practical.

## Development Workflow

Fork the repository and create a branch from the default branch.

Use a descriptive branch name, for example:

```text
feature/add-new-capability
fix/incorrect-validation
docs/update-installation-guide
```

Make your changes and ensure all relevant tests, linting, formatting, and validation checks pass.

Then submit a pull request.

## Commit Messages

Use clear, meaningful commit messages.

Conventional Commits are recommended:

```text
feat: add new capability
fix: correct validation behaviour
docs: update installation instructions
test: add service tests
refactor: simplify repository implementation
chore: update dependencies
ci: update build pipeline
```

Scopes may be used where useful:

```text
feat(api): add health endpoint
fix(terraform): correct provider configuration
docs(helm): document chart values
```

## Code Quality

Contributions should:

- Follow the existing architecture and coding conventions.
- Be easy to understand and maintain.
- Avoid unnecessary complexity.
- Follow SOLID and separation-of-concerns principles where appropriate.
- Include appropriate error handling.
- Avoid introducing unnecessary dependencies.
- Avoid committing generated artifacts unless they are intentionally version controlled.

## Testing

New functionality should include appropriate tests.

Bug fixes should preferably include a regression test demonstrating the original problem.

All existing tests must continue to pass.

## Security

Never commit passwords, API keys, access tokens, private keys, cloud credentials, Terraform state containing sensitive information, or `.env` files containing secrets.

If you accidentally commit a secret, assume the secret has been compromised and rotate or revoke it immediately.

Security vulnerabilities should be reported according to [SECURITY.md](SECURITY.md), not through a public GitHub issue.

## Pull Requests

Pull requests should have a clear title, explain what changed and why, reference related issues where appropriate, include tests for behavioural changes, update documentation where appropriate, and pass all required CI checks.

## License

By contributing to this project, you agree that your contributions will be licensed under the project's MIT License.

## Maintainer

**Steve Owens**  
GitHub: @sowens81  
Email: stevejowens81@outlook.com
