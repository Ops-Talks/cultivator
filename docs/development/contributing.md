# Contributing

Thank you for your interest in contributing to Cultivator. We welcome contributions of all kinds.

## Ways to Contribute

- Report bugs by opening issues
- Request features with clear descriptions
- Submit pull requests with improvements
- Enhance documentation and examples
- Help test and find edge cases

## Getting Started

1. Read the [Development Guide](development.md) to understand project structure
2. Review the [Testing Guide](TESTING.md) for testing strategy and guidelines
3. Set up your local environment (see Development Guide)
4. Pick an issue or feature to work on
5. Submit a pull request

## Pull Request Checklist

- Tests pass locally: `go test ./...`
- Code is formatted: `go fmt ./...`
- Linters pass: `golangci-lint run`
- Tests included for new features
- Documentation updated if needed
- Commit messages are clear and descriptive

## Code Style

Follow the guidelines in the [Development Guide](development.md):

- Use clear, descriptive names
- Document all exported symbols
- Add context to errors using `fmt.Errorf` with `%w` verb
- Keep functions focused and manageable
- Write table-driven tests for multiple cases

## Testing Requirements

New code must include tests. See the [Testing Guide](TESTING.md) for:

- How to write unit tests
- How to add integration tests
- Fuzz testing for robustness
- Coverage expectations

Minimum test coverage for new code: 80% for critical functionality.

## Questions?

- Check existing [Issues](https://github.com/Ops-Talks/cultivator/issues)
- Review the [Development Guide](development.md)
- Open a new issue with your question

---

Thank you for improving Cultivator!
