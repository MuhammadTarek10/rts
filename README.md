# RTS Monorepo

## Developer Tooling (Pre-commit)

This repo uses `lint-staged` with Husky pre-commit hooks.

### Required local tools

- Node.js + npm
- Go (for `gofumpt`)
- .NET SDK (for `dotnet format`)

### Install Go formatter

```bash
go install mvdan.cc/gofumpt@latest
```

Make sure your Go bin directory is on `PATH` (commonly `$(go env GOPATH)/bin`).

### Enable git hooks

```bash
npm install
```

`npm install` runs `prepare` and installs Husky hooks.
