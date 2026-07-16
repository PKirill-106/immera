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
cmd/
    api/

internal/
    auth/
    users/
    documents/
    dictionary/
    translation/
    platform/

migrations/
docs/