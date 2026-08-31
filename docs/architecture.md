# Immera Backend Architecture

## Overview

The Immera backend is implemented as a modular monolith written in Go.

A modular monolith provides clear business boundaries while keeping deployment, local development, transactions, and observability simpler than a distributed microservice architecture.

The application is deployed as one Go API process. The Python translation service will later be deployed separately because it has different runtime and dependency requirements.

## Technology stack

- Go
- PostgreSQL
- pgx/v5 and pgxpool
- go-chi/chi
- SQL migrations
- Docker
- Docker Compose
- Python FastAPI translation service in a later stage

An ORM is not used. Database access is implemented with explicit SQL through pgx.

## Architectural style

The backend follows these principles:

- modular monolith;
- constructor-based dependency injection;
- repository pattern for persistence;
- application services for use-case orchestration;
- HTTP handlers as transport adapters;
- explicit module boundaries;
- no package-level mutable global state.

The design may borrow ideas from Clean Architecture, but it should avoid unnecessary layers and abstractions.

## High-level repository structure

```text
backend/
    cmd/
        api/
    internal/
        auth/
        dictionary/
        document/
        health/
        platform/
        translation/
        user/
    migrations/
docs/
    decisions/
    openapi.yaml
```

Business code is grouped by module under `backend/internal`. Shared HTTP, database, configuration, logging, and security infrastructure belongs in `backend/internal/platform` and must not contain module-specific business rules.

## Authentication and authorization

Access tokens are short-lived HS256 JWTs.

The backend does not persist access tokens. Protected requests are authenticated by:

1. reading `Authorization: Bearer <token>`;
2. validating the JWT signature and expiration;
3. requiring HS256;
4. reading the user UUID from the `sub` claim;
5. storing the authenticated user ID in the request context.

Protected handlers derive the current user exclusively from this request context. Self-service routes use `/users/me` and do not accept an arbitrary user ID from the client.

Refresh tokens are opaque, cryptographically random values. Only their SHA-256 hashes are stored in PostgreSQL. Rotation transactionally removes the old refresh-token session and creates its replacement.

Deleting a user relies on foreign-key cascades to remove their refresh-token sessions, email-verification tokens, and user settings.

## API contract

OpenAPI 3.1 specification in `docs/openapi.yaml` is the source of truth for the public HTTP API.
Swagger UI is exposed at `/docs` for local API exploration, and the raw specification is served at `/openapi.yaml`.
