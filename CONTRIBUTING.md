# Contributing to the free@home API Client

Thank you for your interest in contributing to the free@home API Client! This document outlines the guidelines for contributing code, writing commits, and maintaining code quality.

## 🚀 Repository Setup

1. Clone the repository:
   ```sh
   git clone https://github.com/pgerke/freeathome.git
   cd freeathome
   ```
2. Initialize Go modules:
   ```sh
   go mod tidy
   ```
3. Ensure you have `pre-commit` installed and configured:
   ```sh
   pre-commit install
   ```

## 📝 Commit Guidelines

We use **Commitizen** to enforce conventional commits. Always use `cz commit` instead of `git commit`:

```sh
cz commit
```

Commit messages should follow this format:

```
type(scope): message
```

- **Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`
- **Scope:** The affected module or feature (optional)
- **Message:** A concise description of the change

Examples:

```
feat(logging): add slog-based structured logging in logfmt format
fix(api): resolve authentication issue with token refresh
```

## ✅ Pre-commit Hooks & Code Quality

We use **pre-commit hooks** to enforce coding standards. These hooks run automatically before each commit and include:

- `golangci-lint` for linting
- `gofmt` to ensure proper formatting
- `go mod tidy` to clean up dependencies

Run manually before committing:

```sh
pre-commit run --all-files
```

## 🛠 Logging Standards

All logs should use Go’s `log/slog` package with the **logfmt** format. Example:

```go
package main
import (
	"log/slog"
	"os"
)
func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("application started", "version", "0.1.0")
}
```

## 📖 Documentation & README Updates

- Any changes to functionality must include relevant documentation updates.
- If you introduce a new feature, update the `README.md` with usage instructions.

## 🔄 Pull Request Process

1. **Fork the repository** and create a feature branch:
   ```sh
   git checkout -b feature/your-feature
   ```
2. **Make your changes**, ensuring:
   - Code follows the style guidelines.
   - Pre-commit hooks pass.
3. **Test your changes locally** before pushing.
4. **Push your branch** and create a pull request.

---

Thank you for contributing! 🎉
