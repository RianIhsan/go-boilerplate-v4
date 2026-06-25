FROM golang:1.25-alpine AS builder

WORKDIR /app

# build-base provides gcc/musl-dev, needed because the DMG metadata parser
# pulls in a cgo dependency (lzfse decompression) — the rest of the app has
# no cgo requirement.
RUN apk add --no-cache git build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o bin/api ./cmd/api/main.go

# ── Final image ───────────────────────────────────────────
FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/bin/api .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./api"]
