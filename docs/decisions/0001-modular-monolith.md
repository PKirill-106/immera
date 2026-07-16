# ADR-0001: Use a Modular Monolith

- Status: Accepted
- Date: 2026-07-16
- Decision owners: Immera engineering

## Context

Immera requires several distinct business areas:

- authentication;
- users;
- documents and reading progress;
- contextual translation;
- personal dictionary;
- future notes and bookmarks.

The system may later integrate with external translation providers and a Python service that handles LLM structured outputs.

A traditional unstructured monolith would be simple to start but would allow business concerns to become tightly coupled.

A microservice architecture would provide strong deployment boundaries, but it would also introduce:

- network communication between business areas;
- distributed transactions;
- additional deployment and observability complexity;
- contract versioning;
- more difficult local development;
- operational overhead that is not justified by the current scale or team size.

## Decision

The Go backend will be implemented as a modular monolith.

The application will run as one deployable Go API process while keeping explicit boundaries between business modules.

Initial modules include:

- Auth;
- Users;
- Documents;
- Translation;
- Dictionary.

Technical infrastructure will live under `internal/platform`.

Each module owns its domain types, services, persistence contracts, repository implementation, HTTP handlers, DTOs, and route registration.

Modules must not directly depend on another module's concrete repository.

Cross-module interactions should use deliberately defined interfaces or application services.

The Python translation service is allowed to remain a separate deployable component because it uses a different language, runtime, libraries, and scaling profile.

## Consequences

### Positive

- one deployable Go backend;
- straightforward local development;
- simple transactions inside PostgreSQL;
- lower operational cost;
- clear business boundaries;
- easier refactoring than an unstructured monolith;
- modules can be extracted later if justified.

### Negative

- module boundaries are enforced primarily through code organization and engineering discipline;
- all Go modules share one process and deployment lifecycle;
- one poorly designed module can still affect the entire application;
- database ownership boundaries are less strict than with isolated databases;
- careless imports can create coupling or dependency cycles.

## Rules resulting from this decision

- Business code is organized by module, not by global technical layer.
- There is no repository-wide `handlers`, `services`, or `repositories` directory.
- Shared platform code must not contain business rules.
- Module constructors receive dependencies explicitly.
- Modules may not import another module's PostgreSQL implementation.
- New shared abstractions require more than one concrete use case.
- The system will not be split into microservices only for organizational appearance.
- Extraction requires measurable reasons such as independent scaling, security isolation, ownership, runtime differences, or deployment constraints.

## Alternatives considered

### Layered monolith

Example:

```text
handlers/
services/
repositories/
models/