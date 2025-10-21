FROM golang:1.23-bookworm AS builder
WORKDIR /src

# Download dependencies first for better build cache reuse
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source and build a static binary for Linux/amd64
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /src/bin/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

# Copy the compiled binary (and migrations if they are needed at runtime)
COPY --from=builder /src/bin/api ./api
COPY --from=builder /src/migrations ./migrations

ENV APP_HOST=0.0.0.0
ENV APP_PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/api"]
