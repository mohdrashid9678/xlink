# xlink

A highly available and efficient URL shortener.

## URL API

Start PostgreSQL, create a `.env` with `DB_URL`, apply the initial schema with
`make db-init`, then start the API with `make run`.

| Method   | Path                      | Purpose                |
| -------- | ------------------------- | ---------------------- |
| `POST`   | `/api/v1/urls`            | Create a short URL     |
| `GET`    | `/api/v1/urls/:shortCode` | Retrieve one short URL |
| `PATCH`  | `/api/v1/urls/:shortCode` | Update approved fields |
| `DELETE` | `/api/v1/urls/:shortCode` | Delete a short URL     |

Example create request:

```json
{
  "long_url": "https://example.com/docs",
  "custom_alias": "docs",
  "expires_at": "2027-01-01T00:00:00Z"
}
```
