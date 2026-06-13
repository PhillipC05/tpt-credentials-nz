# =============================================================================
# Stage 1: Build the Go backend binary
# =============================================================================
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Cache module downloads
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/server \
    ./cmd/server

# =============================================================================
# Stage 2: Build the Next.js frontend
# =============================================================================
FROM node:20-alpine AS frontend-builder

WORKDIR /web

COPY web/package.json web/ ./
RUN npm ci

COPY web/ .

RUN npm run build

# =============================================================================
# Stage 3: Minimal runtime image
# =============================================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=frontend-builder /web/.next .next
COPY --from=frontend-builder /web/public public
COPY --from=frontend-builder /web/package.json .
COPY --from=frontend-builder /web/node_modules ./node_modules

EXPOSE 8094

CMD ["/app/server"]