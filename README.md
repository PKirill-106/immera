# Immera

📚 Read foreign-language books without breaking your reading flow.

Immera is a modular-monolith Go backend for an e-reading and vocabulary-building application. PostgreSQL is its primary datastore.

## Development

Copy `.env.example` to `.env`, then run the complete stack:

```sh
make docker-up
```

For local API development, start PostgreSQL and run the API separately:

```sh
docker compose up -d postgres
make run
```

Useful checks:

```sh
make fmt
make test
make vet
```

The API exposes `GET /health/live` for liveness and `GET /health/ready` for PostgreSQL-backed readiness.
