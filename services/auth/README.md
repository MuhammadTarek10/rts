# Auth Service

The Auth service is the identity and access management component of the RTS platform. It handles user registration, login, session management, and exposes user profile management endpoints. It is built with NestJS (TypeScript), persists data in PostgreSQL via Drizzle ORM, and publishes auth events to RabbitMQ for consumption by other services.

## Tech Stack

| Concern | Technology |
|---|---|
| Framework | NestJS 11 |
| Language | TypeScript |
| Database | PostgreSQL (via `postgres` driver) |
| ORM | Drizzle ORM |
| Auth | Passport.js (`passport-local`, `passport-jwt`) |
| Password hashing | Argon2 |
| Message broker | RabbitMQ (AMQP via `@nestjs/microservices`) |
| API docs | Swagger / OpenAPI (password-protected) |
| Validation | `class-validator` + `zod` (env schema) |

---

## Architecture

The service is organised into a root `AppModule` that composes four core modules and two feature modules.

```
src/
├── main.ts                          # Bootstrap: global prefix, versioning, CORS, Swagger, cookies
├── app.module.ts                    # Root module
├── common/
│   ├── constants.ts                 # Shared enums/constants (strategies, roles, statuses, broker keys)
│   └── dtos/response.dto.ts         # Generic ResponseDto wrapper used by all endpoints
├── core/
│   ├── config/                      # NestJS ConfigModule namespace configs (app, database, jwt, broker, swagger)
│   │   └── validation.schema.ts     # Zod schema validating all required env vars at startup
│   ├── database/
│   │   ├── database.module.ts       # Provides a Drizzle client bound to the DATABASE injection token
│   │   ├── base.repository.ts       # BaseRepository with transaction and query helpers
│   │   ├── types.ts                 # PostgresDatabase / PostgresTransaction type aliases
│   │   └── schemas/                 # Drizzle table definitions and relations (see Database Schema)
│   ├── broker/
│   │   └── broker.module.ts         # Global module: registers RabbitMQ ClientProxy for AUTH_EVENTS queue
│   ├── utils/
│   │   ├── utils.module.ts          # Exports TokenService, CookieService, HashService, RandomService
│   │   ├── token/token.service.ts   # JWT generation and verification (access + refresh)
│   │   ├── cookie/cookie.service.ts # Sets / clears HTTP-only auth cookies
│   │   ├── hash/hash.service.ts     # Argon2 hash and verify
│   │   └── random/random.service.ts # UUID generation
│   ├── decorators/                  # @CurrentUser, @RefreshToken, @ResponseMessage
│   ├── interceptors/
│   │   └── response.interceptor.ts  # Wraps every response in ResponseDto envelope
│   └── filters/
│       └── http-exception.ts        # (stub, not yet active)
├── auth/
│   ├── auth.module.ts               # Feature module: strategies, repositories, publisher
│   ├── auth.controller.ts           # Routes: sign-up, sign-in, refresh, sign-out
│   ├── auth.service.ts              # Business logic: validate, signUp, signIn, refresh, logout
│   ├── auth.repository.ts           # DB queries: create user, find by email/id, lock/unlock account
│   ├── session.repository.ts        # DB queries: create, find, update, delete sessions
│   ├── auth-events.publisher.ts     # Emits UserRegistered event to RabbitMQ
│   ├── strategies/
│   │   ├── local.strategy.ts        # Passport local — validates email + password credentials
│   │   ├── jwt.strategy.ts          # Passport JWT — validates access token (cookie or Bearer)
│   │   └── refresh-jwt.strategy.ts  # Passport JWT — validates refresh token (cookie or Bearer)
│   ├── guards/
│   │   ├── local.guard.ts
│   │   ├── jwt.guard.ts
│   │   └── refresh-jwt.guard.ts
│   └── dtos/
│       ├── sign-in.dto.ts
│       ├── sign-up.dto.ts
│       └── token-response.dto.ts
└── users/
    ├── users.module.ts              # Feature module: profile and account management
    ├── users.controller.ts          # Routes: get/update profile, delete account, change password
    ├── users.service.ts             # Business logic for profile and password operations
    ├── user.repository.ts           # DB queries: find profile, soft-delete, session cleanup
    └── dtos/
        ├── profile-response.dto.ts
        ├── update-profile.dto.ts
        └── change-password.dto.ts
```

---

## API Endpoints

All routes are prefixed with `/api/v1`. The global `ResponseInterceptor` wraps every successful response in the following envelope:

```json
{
  "data": { ... },
  "message": "Human-readable message",
  "status": "success"
}
```

### Authentication (`/api/v1`)

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/sign-up` | None | Register a new user |
| `POST` | `/sign-in` | None (LocalGuard validates credentials) | Authenticate and receive tokens |
| `POST` | `/refresh` | Refresh token (cookie or Bearer) | Issue new access + refresh token pair |
| `POST` | `/sign-out` | Access token (cookie or Bearer) | Invalidate current session and clear cookies |

#### POST /sign-up

Request body:

```json
{
  "email": "user@example.com",
  "password": "P@ssw0rd",
  "first_name": "John",
  "last_name": "Doe",
  "phone_number": "+1234567890",
  "country": "USA",
  "date_of_birth": "1990-01-01"
}
```

`email` and `password` are required. All other fields are optional. Password rules: minimum 8 characters, must contain at least one lowercase letter, one uppercase letter, one digit, and one special character (`@$!%*?&`).

Response `201`:

```json
{
  "data": {
    "access_token": "<jwt>",
    "refresh_token": "<jwt>",
    "expires_in": 3600
  },
  "message": "User signed up successfully",
  "status": "success"
}
```

Tokens are also set as HTTP-only cookies (`access_token`, `refresh_token`).

#### POST /sign-in

Request body:

```json
{
  "email": "user@example.com",
  "password": "P@ssw0rd"
}
```

Response `200`: same shape as sign-up. Returns `401` if credentials are invalid or the account is locked.

#### POST /refresh

Provide the refresh token either as an HTTP-only cookie (`refresh_token`) or as a `Bearer` token in the `Authorization` header.

Response `200`: same token response shape with a new token pair.

#### POST /sign-out

Provide the access token either as an HTTP-only cookie (`access_token`) or as a `Bearer` token.

Response `200`: clears `access_token` and `refresh_token` cookies. `data` is `null`.

---

### Identity / Users (`/api/v1/users`)

All endpoints in this group require the access token (cookie or Bearer).

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/users/profile` | JWT | Return authenticated user's profile |
| `PATCH` | `/users/profile` | JWT | Partially update the authenticated user's profile |
| `DELETE` | `/users/me` | JWT | Soft-delete the authenticated user's account and revoke all sessions |
| `PATCH` | `/users/change-password` | JWT | Change password and invalidate all sessions |

#### GET /users/profile

Response `200`:

```json
{
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "status": "active",
    "role": "user",
    "profile": {
      "id": "uuid",
      "first_name": "John",
      "last_name": "Doe",
      "avatar_url": null,
      "phone_number": null,
      "date_of_birth": null,
      "country": null,
      "bio": null
    },
    "created_at": "2024-01-01T00:00:00.000Z",
    "updated_at": "2024-01-01T00:00:00.000Z"
  },
  "message": "Profile retrieved successfully",
  "status": "success"
}
```

#### PATCH /users/profile

All fields optional:

```json
{
  "first_name": "John",
  "last_name": "Doe",
  "phone_number": "+1234567890",
  "country": "USA",
  "date_of_birth": "1990-01-01",
  "bio": "Software engineer.",
  "avatar_url": "https://example.com/avatar.png"
}
```

#### DELETE /users/me

Response `204 No Content`. Soft-sets `deleted_at` on the `users` row and deletes all session records.

#### PATCH /users/change-password

```json
{
  "current_password": "OldP@ssw0rd",
  "new_password": "NewP@ssw0rd1"
}
```

Response `200`. All active sessions are deleted after the password is updated.

---

## Authentication Flows

### Local sign-in

1. Client `POST /sign-in` with `{ email, password }`.
2. `LocalGuard` triggers `LocalStrategy.validate`, which calls `AuthService.validateUser`.
   - Checks whether the account is locked (`locked_until` field in `auths`).
   - Verifies the Argon2 password hash.
   - On failure, increments `failed_attempts`; locks the account for 15 minutes after 5 consecutive failures (`LOCKDOWN_THRESHOLD = 5`).
   - On success, resets `failed_attempts`.
3. `AuthService.signIn` generates a UUID session ID, issues an access token and refresh token (both containing `{ id, email, session_id }`), stores a hashed refresh token in the `sessions` table, and returns the token pair.
4. `CookieService.setAuthCookies` writes both tokens as HTTP-only cookies (`SameSite=Lax`, `Secure` in production only).

### Registration

1. Client `POST /sign-up` with user details.
2. `AuthService.signUp` runs a database transaction:
   - Checks for email uniqueness.
   - Inserts into `users` (status = `pending`), `user_profiles`, and `auths` (strategy = `local`).
   - Generates tokens and creates a session.
3. After the transaction commits, `AuthEventsPublisher.emitUserRegistered` fires a `UserRegistered` event to RabbitMQ (fire-and-forget; failures are logged as warnings but do not fail the request).

### Token refresh

1. Client `POST /refresh` with the refresh token (cookie or Bearer).
2. `RefreshGuard` triggers `RefreshTokenStrategy.validate`, which extracts the raw token from the cookie or `Authorization` header and attaches it to the payload.
3. `AuthService.refresh` runs a transaction:
   - Looks up the session by `session_id` from the token payload.
   - Verifies the raw refresh token against the stored Argon2 hash.
   - Issues a new token pair and updates the session hash and expiry (token rotation).

### Sign-out

1. Client `POST /sign-out` with the access token.
2. `JwtGuard` validates the token; `session_id` is extracted from the JWT payload via the `@CurrentUser` decorator.
3. `AuthService.logout` deletes the specific session row (or all sessions for the user if no `session_id` is present).
4. `CookieService.clearAuthCookies` expires both cookies immediately.

### JWT token payload

Both the access token and refresh token carry:

```json
{
  "id": "<user uuid>",
  "email": "user@example.com",
  "session_id": "<session uuid>",
  "jti": "<unique token id>",
  "iat": 1700000000,
  "exp": 1700003600
}
```

The access token is signed with `JWT_ACCESS_SECRET`; the refresh token is signed with a separate `JWT_REFRESH_SECRET`. Both Passport strategies accept the token from an HTTP-only cookie first, falling back to the `Authorization: Bearer` header.

---

## Environment Variables

The service validates all environment variables at startup using Zod. The application will refuse to start if a required variable is missing or invalid.

| Variable | Required | Default | Description |
|---|---|---|---|
| `NODE_ENV` | No | `development` | Runtime environment (`development`, `production`, `test`, `staging`) |
| `PORT` | No | `8001` | HTTP port the service listens on |
| `FRONTEND_URL` | No | `http://localhost:3000` | Allowed CORS origin |
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `JWT_ACCESS_SECRET` | Yes | — | Secret for signing access tokens |
| `JWT_ACCESS_EXPIRATION` | No | `86400` | Access token TTL in seconds (default: 1 day) |
| `JWT_REFRESH_SECRET` | Yes | — | Secret for signing refresh tokens |
| `JWT_REFRESH_EXPIRATION` | No | `604800` | Refresh token TTL in seconds (default: 7 days) |
| `SWAGGER_USERNAME` | Yes | — | Basic-auth username to access `/docs` |
| `SWAGGER_PASSWORD` | Yes | — | Basic-auth password to access `/docs` |
| `RABBITMQ_URI` | Yes | — | RabbitMQ AMQP connection URI |

Create `.env.development` (or `.env.production`) in `services/auth/`:

```dotenv
# App
NODE_ENV=development
PORT=8001
FRONTEND_URL=http://localhost:3000

# Database
DATABASE_URL=postgresql://postgres:postgres@localhost:5439/auth_db

# JWT
JWT_ACCESS_SECRET=your_jwt_secret_key
JWT_ACCESS_EXPIRATION=3600
JWT_REFRESH_SECRET=your_jwt_refresh_secret_key
JWT_REFRESH_EXPIRATION=86400

# Swagger
SWAGGER_USERNAME=admin
SWAGGER_PASSWORD=admin

# RabbitMQ
RABBITMQ_URI=amqp://guest:guest@localhost:5672
```

---

## Development Setup

### Prerequisites

- Node.js (LTS)
- Docker (for the local PostgreSQL instance)

### Install dependencies

```bash
cd services/auth
npm install
```

### Start the local database

The service ships a `docker-compose.yml` that starts a `postgres:18-alpine` container on port `5439` with database `auth_db`.

```bash
cd services/auth
docker compose up -d
```

### Run migrations

```bash
# Generates SQL migration files from the Drizzle schema, then applies them
npm run dev:migrate
```

Migration files are stored at `src/core/database/migrations/` and tracked in a `migrations` table in PostgreSQL.

### Start the service

```bash
npm run dev      # watch mode with NODE_ENV=development
```

The API is available at `http://localhost:8001/api/v1`.

Swagger UI is available at `http://localhost:8001/docs` (protected by basic auth using `SWAGGER_USERNAME` / `SWAGGER_PASSWORD`).

### Drizzle Studio (database GUI)

```bash
npm run dev:studio
```

### Tests

```bash
npm run test          # unit tests
npm run test:e2e      # end-to-end tests
npm run test:cov      # unit tests with coverage report
```

---

## Database Schema

The schema is defined with Drizzle ORM in `src/core/database/schemas/`.

### `users`

| Column | Type | Notes |
|---|---|---|
| `id` | `uuid` PK | Auto-generated |
| `email` | `text` | Unique, not null |
| `status` | `user_status` enum | `pending` \| `active` \| `inactive` \| `suspended`; default `active` |
| `role` | `user_roles` enum | `user` \| `admin`; default `user` |
| `email_verified_at` | `timestamp` | Nullable |
| `last_login_at` | `timestamp` | Nullable |
| `password_changed_at` | `timestamp` | Nullable |
| `created_at` | `timestamp` | Default `now()` |
| `updated_at` | `timestamp` | Default `now()`, auto-updated on change |
| `deleted_at` | `timestamp` | Nullable; set on soft-delete |

Relations: `hasMany auths`, `hasOne userProfile`.

### `auths`

Holds one row per authentication method per user. Currently only the `local` strategy is implemented; the schema supports `google`, `github`, and `facebook` for future OAuth strategies.

| Column | Type | Notes |
|---|---|---|
| `id` | `uuid` PK | Auto-generated |
| `user_id` | `uuid` FK | References `users.id` (cascade delete) |
| `strategy` | `auth_strategy` enum | `local` \| `google` \| `github` \| `facebook` |
| `provider_user_id` | `text` | Nullable; for OAuth strategies |
| `is_primary` | `boolean` | Default `false` |
| `password_hash` | `text` | Nullable; Argon2 hash (local strategy only) |
| `failed_attempts` | `integer` | Default `0`; incremented on wrong password |
| `locked_until` | `timestamp` | Nullable; set when `failed_attempts` >= 5 |
| `last_used_at` | `timestamp` | Nullable |
| `created_at` | `timestamp` | Default `now()` |
| `updated_at` | `timestamp` | Auto-updated |
| `deleted_at` | `timestamp` | Nullable |

Indexes: `(strategy, user_id)`, `(strategy, provider_user_id)`, `(locked_until)`.

### `sessions`

Stores active refresh token sessions. A single user can have multiple concurrent sessions (one per sign-in from a different client).

| Column | Type | Notes |
|---|---|---|
| `id` | `uuid` PK | Provided by service (random UUID) |
| `auth_id` | `uuid` FK | References `auths.id` (cascade delete) |
| `refresh_token_hash` | `text` | Argon2 hash of the issued refresh token |
| `expires_at` | `timestamp` | Derived from `JWT_REFRESH_EXPIRATION` |
| `revoked_at` | `timestamp` | Nullable |
| `ip_address` | `text` | Nullable |
| `user_agent` | `text` | Nullable |
| `device_id` | `text` | Nullable |
| `created_at` | `timestamp` | Default `now()` |
| `deleted_at` | `timestamp` | Nullable |

Indexes: `(auth_id)`, `(refresh_token_hash)`, `(expires_at)`, `(revoked_at)`.

### `user_profiles`

| Column | Type | Notes |
|---|---|---|
| `id` | `uuid` PK | Auto-generated |
| `user_id` | `uuid` FK | References `users.id` (cascade delete) |
| `first_name` | `text` | Nullable |
| `last_name` | `text` | Nullable |
| `avatar_url` | `text` | Nullable |
| `phone_number` | `text` | Nullable |
| `date_of_birth` | `date` | Nullable |
| `country` | `text` | Nullable |
| `bio` | `text` | Nullable |
| `created_at` | `timestamp` | Default `now()` |
| `updated_at` | `timestamp` | Auto-updated |

Unique index: `(user_id)`.

---

## RabbitMQ Events

The service publishes to the `auth.events` queue (durable).

| Event pattern | Payload | Trigger |
|---|---|---|
| `UserRegistered` | `{ id: string, email: string }` | Successful `POST /sign-up` |

Events are emitted fire-and-forget. If the RabbitMQ connection is unavailable the failure is logged as a warning and the HTTP response is not affected.

Other services that need to react to user registration should consume the `auth.events` queue and filter for the `UserRegistered` pattern.
