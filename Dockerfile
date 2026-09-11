FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /nexrouter ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates wget
WORKDIR /app
COPY --from=builder /nexrouter /app/nexrouter
RUN mkdir -p /app/data
ENV DB_PATH=/app/data/nexrouter.db
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 CMD wget -qO- http://localhost:8080/health || exit 1
ENTRYPOINT ["/app/nexrouter"]