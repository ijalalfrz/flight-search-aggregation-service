FROM golang:1.25-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
# CGO_ENABLED=0 produces a statically linked binary which is required for distroless/static images
# -ldflags="-s -w" reduces binary size by stripping debug information
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o service ./cmd/main.go

FROM gcr.io/distroless/static-debian12:nonroot

# Set the working directory
WORKDIR /

# Copy the compiled binary from the builder stage
COPY --from=builder /app/service /app

COPY --from=builder /app/tests /tests

# Expose the default port (update this if your HTTP_PORT is different)
EXPOSE 8080

# Explicit set nonroot user even though already nonroot to makesure the user
USER nonroot

ENTRYPOINT ["/app"]