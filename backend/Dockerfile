# Stage 1: Build binary
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency definitions
COPY go.mod go.sum* ./
RUN go mod download || true

# Copy source code
COPY . .

# Build statically linked Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /choresync-backend ./cmd/server

# Stage 2: Minimal runtime image
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /choresync-backend /app/choresync-backend

ENV PORT=8000
EXPOSE 8000

USER nobody:nobody

ENTRYPOINT ["/app/choresync-backend"]
