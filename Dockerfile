# Multi-stage build for minimal production container size & security
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/xlink ./cmd/api

FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/bin/xlink /app/xlink

EXPOSE 8080

USER nobody:nobody

ENTRYPOINT ["/app/xlink"]
