# Immera Database Schema

## Status

Draft.

This document describes the intended data model and important invariants.

SQL migrations are the executable source of truth. When this document conflicts with applied migrations, migrations take precedence and this document should be updated.

## General conventions

- PostgreSQL is the primary database.
- Primary keys use UUID values.
- Database column names use `snake_case`.
- Timestamps use `TIMESTAMPTZ`.
- Application code should store and compare timestamps in UTC.
- Foreign keys are used to enforce referential integrity.
- Required fields use `NOT NULL`.
- User-owned data must always include an ownership relationship.
- Passwords and raw refresh tokens must never be stored in plaintext.

The exact UUID version will be decided before the first production migration.

## Module ownership

| Module      | Tables                                    |
| ----------- | ----------------------------------------- |
| Auth        | `users`, `auth_refresh_tokens`, `email_verification_tokens` |
| Documents   | `documents`, `document_progress`          |
| Translation | `global_words`, `translation_cache`       |
| Dictionary  | `dictionary_entries`, `dictionary_groups` |

Table names may still be refined before migrations are finalized.

## Users

### `users`

Represents an Immera account.

Proposed fields:

| Column          | Type        | Constraints      |
| --------------- | ----------- | ---------------- |
| `unique_id`     | UUID        | primary key      |
| `email`         | TEXT        | unique, not null |
| `password_hash` | TEXT        | not null         |
| `created_at`    | TIMESTAMPTZ | not null         |
| `email_verified_at` | TIMESTAMPTZ | nullable     |

Important rules:

- email comparison must be case-insensitive;
- password hashes are created using a standard password-hashing algorithm;
- raw passwords are never persisted.

Open decision:

- use normalized email in a separate column;
- use PostgreSQL `citext`;
- or enforce lowercase email in application and database constraints.

## Refresh tokens

### `auth_refresh_tokens`

Stores server-side refresh-token sessions.

Proposed fields:

| Column       | Type        | Constraints      |
| ------------ | ----------- | ---------------- |
| `id`         | UUID        | primary key      |
| `user_id`    | UUID        | FK to `users`    |
| `token_hash` | TEXT        | unique, not null |
| `expires_at` | TIMESTAMPTZ | not null         |
| `created_at` | TIMESTAMPTZ | not null         |

Important rules:

- only a cryptographic hash of the refresh token is stored;
- expired and rotated or deleted tokens cannot be used;
- deleting a user should remove their refresh-token sessions;
- token rotation should be supported by the application service.
- rotation replaces the stored session atomically; a separate revocation timestamp is not currently used.

Likely indexes:

- unique index on `token_hash`;
- index on `user_id`;
- optional index on `expires_at` for cleanup jobs.

## Email verification tokens

### `email_verification_tokens`

Stores one-time email-verification tokens. Only SHA-256 hashes are persisted.

| Column       | Type        | Constraints                    |
| ------------ | ----------- | ------------------------------ |
| `id`         | UUID        | primary key                    |
| `user_id`    | UUID        | FK to `users`, not null        |
| `token_hash` | VARCHAR(64) | unique, not null               |
| `expires_at` | TIMESTAMPTZ | not null                       |
| `created_at` | TIMESTAMPTZ | not null                       |
| `used_at`    | TIMESTAMPTZ | nullable                       |

Important rules:

- raw email-verification tokens are never stored;
- verification tokens expire and can only be used once;
- setting `users.email_verified_at` and marking the token used happen atomically;
- resending replaces previous unused verification tokens for the user;
- deleting a user removes their verification tokens.

## Documents

### `documents`

Represents a document owned by a user.

Proposed fields:

| Column            | Type        | Constraints                            |
| ----------------- | ----------- | -------------------------------------- |
| `unique_id`       | UUID        | primary key                            |
| `user_id`         | UUID        | FK to `users`, not null                |
| `title`           | TEXT        | not null                               |
| `author`          | TEXT        | nullable                               |
| `source_language` | VARCHAR     | nullable initially                     |
| `target_language` | VARCHAR     | nullable                               |
| `format`          | VARCHAR     | not null                               |
| `file_url`        | TEXT        | nullable depending on storage strategy |
| `cover_url`       | TEXT        | nullable                               |
| `created_at`      | TIMESTAMPTZ | not null                               |
| `updated_at`      | TIMESTAMPTZ | not null                               |

The original draft stored document content and current position directly in this table. The recommended direction is:

- keep metadata in `documents`;
- store reading position separately;
- define the storage strategy for extracted document content before finalizing the schema.

Likely indexes:

- index on `(user_id, created_at DESC)`;
- optional uniqueness constraint based on a future upload checksum.

### `document_progress`

Stores a user's reading progress for one document.

Proposed fields:

| Column            | Type          | Constraints                 |
| ----------------- | ------------- | --------------------------- |
| `unique_id`       | UUID          | primary key                 |
| `user_id`         | UUID          | FK to `users`, not null     |
| `document_id`     | UUID          | FK to `documents`, not null |
| `current_locator` | TEXT or JSONB | not null                    |
| `percentage`      | NUMERIC       | nullable                    |
| `last_opened_at`  | TIMESTAMPTZ   | not null                    |
| `created_at`      | TIMESTAMPTZ   | not null                    |
| `updated_at`      | TIMESTAMPTZ   | not null                    |

Important rules:

- one progress record per user and document;
- `percentage`, when present, must be between 0 and 100;
- progress must belong to the same user who owns or can access the document.

Required constraint:

```text
UNIQUE (user_id, document_id)
```
