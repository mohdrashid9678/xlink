# xlink

A Go URL shortener implementing the document's first two core requirements:

1. **Phase 1 - create links:** `POST /api/v1/shorten` validates a destination, creates a collision-safe Base62 code (or an atomic custom alias), and records optional expiration.
2. **Phase 2 - redirect links:** `GET /{short_code}` resolves a live link and returns a `302 Found`; missing and expired links return `404` and `410` respectively.

The current repository uses a concurrency-safe in-memory store so it runs without infrastructure. The `link.Store` interface is the seam for a durable database and Redis-backed read cache in the next phase.

## Run

```sh
go run .
```

Set `BASE_URL` when the public address differs from `http://localhost:8080`.

## Create a link

```sh
curl -i http://localhost:8080/api/v1/shorten \
  -H 'Content-Type: application/json' \
  -d '{"long_url":"https://example.com/products","custom_alias":"summer-sale"}'
```

```json
{
  "short_url": "http://localhost:8080/summer-sale",
  "long_url": "https://example.com/products",
  "created_at": "2026-08-03T00:00:00Z"
}
```

`custom_alias` is optional and must be 3-50 URL-safe letters, numbers, hyphens, or underscores. `expires_at` is optional RFC3339, for example `2026-12-31T23:59:59Z`.

## Test

```sh
go test ./...
```
