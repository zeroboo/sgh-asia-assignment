# ---- Build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary.
COPY . .
RUN CGO_ENABLED=0 go build -o /payment-service ./server

# ---- Runtime stage ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /payment-service /payment-service

ENV GIN_MODE=release
EXPOSE 8080

ENTRYPOINT ["/payment-service"]
