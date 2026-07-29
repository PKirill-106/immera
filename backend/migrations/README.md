# Database migrations

This directory contains the ordered PostgreSQL schema history for the Immera
backend. Migrations are managed with Goose and use timestamp versions.

Create a migration from the repository root:

```sh
make migrate-create name=create_documents
```

Each SQL migration contains a forward and reverse section:

```sql
-- +goose Up
-- SQL that advances the schema.

-- +goose Down
-- SQL that reverses the change.
```

Once a migration has been applied to a shared environment, do not edit or
renumber it. Add a new migration that corrects or evolves the schema instead.
