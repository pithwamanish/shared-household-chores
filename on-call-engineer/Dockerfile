# Multi-stage Go build for ChoreSync On-Call Alert Receiver
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY receiver.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/receiver receiver.go

# Production minimal runtime image
FROM alpine:3.19

RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=builder /app/bin/receiver /app/receiver
COPY payload-schema.json /app/payload-schema.json
COPY autonomy-policy.json /app/autonomy-policy.json

EXPOSE 5050

CMD ["/app/receiver"]
