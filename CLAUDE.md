# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

RTS is a microservices monorepo with a polyglot backend and a Vue 3 frontend:

| Service | Language/Framework | Database | Port |
|---|---|---|---|
| `services/auth` | NestJS (TypeScript) | PostgreSQL (Drizzle ORM) | 3001 |
| `services/catalog` | ASP.NET Core (.NET 10) | MongoDB | - |
| `services/inventory` | (planned) | - | - |
| `services/orders` | (planned) | - | - |
| `services/payment` | (planned) | - | - |
| `services/notification` | (planned) | - | - |
| `services/recommendation` | (planned) | - | - |
| `app` | Vue 3 + Vite + Pinia | - | - |

**Infrastructure (docker-compose.yml):** PostgreSQL, MongoDB, Redis, RabbitMQ, MinIO (S3-compatible), Nginx, Grafana, Certbot.

## Auth Service (`services/auth`)

NestJS app with Passport.js strategies (local, JWT access token, JWT refresh token). Auth events are published to RabbitMQ for other services to consume.

**Module structure:** `core/` (config, database, broker, utils, decorators, filters, interceptors) + `auth/` + `users/`

**Migrations:** Drizzle Kit, migrations stored at `src/core/database/migrations/`.

```bash
# Dev
cd services/auth
npm install
npm run dev              # watch mode with NODE_ENV=development

# Migrations (dev)
npm run dev:migrate      # generate + migrate using .env.development

# Tests
npm run test             # unit tests
npm run test:e2e         # e2e tests
npm run test:cov         # coverage

# Local dev DB
docker compose up -d     # starts auth-database (postgres:18 on port 5439)

# Drizzle Studio
npm run dev:studio
```

## Catalog Service (`services/catalog`)

ASP.NET Core Web API (.NET 10) using Clean Architecture layers: `Domain/`, `Application/`, `Infrastructure/`, `Controllers/`. JWT bearer auth is validated using the shared secret from the auth service.

```bash
# Dev
cd services/catalog
docker compose up -d            # starts MongoDB on port 27020

cd Catalog.Api
dotnet run                      # starts the API (Swagger at /swagger in dev)
dotnet test                     # run tests (solution-level)
dotnet format Catalog.slnx      # format C# code
```

## Frontend (`app`)

Vue 3 + Vite + Pinia (with persisted state) + Vue Router. Written in TypeScript, styled with LESS.

```bash
cd app
npm install
npm run dev      # vite dev server
npm run build    # vue-tsc + vite build
```

## Full Stack (docker-compose)

```bash
# Start all infrastructure + services
docker compose up -d

# Start individual service
docker compose up -d db mongo redis rabbitmq
```

## Pre-commit Hooks

Root-level Husky + lint-staged runs automatically on commit:
- TypeScript files → ESLint (auth + app)
- Vue files → ESLint
- Go files → `gofumpt`
- C# files → `dotnet format services/catalog/Catalog.slnx --include`

Setup once per machine:
```bash
npm install   # from repo root — installs Husky hooks
```

Required tools on PATH: `node`, `go` (with `gofumpt` installed), `.NET SDK`.

Install gofumpt: `go install mvdan.cc/gofumpt@latest`

## Commit Convention

Conventional commits enforced via `commitlint` (e.g. `feat:`, `fix:`, `chore:`).

## gstack

Use the `/browse` skill from gstack for all web browsing. Never use `mcp__chrome-devtools__*` tools.

**Available skills:**
- `/plan-ceo-review` - CEO-level planning review
- `/plan-eng-review` - Engineering review planning
- `/review` - Code review
- `/ship` - Ship/deploy functionality
- `/browse` - Fast headless browser for testing and QA
- `/qa` - QA testing
- `/setup-browser-cookies` - Setup browser cookies
- `/retro` - Retro/review functionality
