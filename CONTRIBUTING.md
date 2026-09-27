# Contributing to Boa

Thank you for your interest in contributing to **Boa**! We welcome contributions from developers of all skill levels.

---

## Code of Conduct

Please review and adhere to our [Code of Conduct](./CODE_OF_CONDUCT.md) in all project interactions.

---

## How to Contribute

### 1. Reporting Bugs
- Search existing issues to ensure the defect has not already been reported.
- Open a new issue with a clear reproduction case, expected behavior, and environment details (`bo version`).

### 2. Suggesting Features
- Open an issue describing the use case and proposed CLI UX.
- Keep in mind Boa's core philosophy: **dense, terminal-native, and safe by default**.

### 3. Submitting Pull Requests
1. Fork the repository on GitHub.
2. Clone your fork locally:
   ```bash
   git clone https://github.com/<your-username>/Boa.git
   cd Boa
   ```
3. Create a descriptive feature branch:
   ```bash
   git checkout -b feat/my-improvement
   ```
4. Ensure all unit tests and linter pass:
   ```bash
   make test
   make lint
   ```
5. Commit your changes following clean Git conventions:
   ```bash
   git commit -m "feat: add support for custom glob exclusions"
   ```
6. Push to your fork and submit a Pull Request against `main`.

---

## Critical Safety Rules

Any PR modifying `internal/safety/` must preserve all Zip-Slip and path traversal defenses and maintain 100% test coverage in `internal/safety/safety_test.go` and `internal/extract/extract_test.go`.
