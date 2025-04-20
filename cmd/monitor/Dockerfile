# === Build Stage ===
FROM golang:1.24 AS builder

ARG commit=unknown
ARG version=unknown

WORKDIR /app

# Download modules if neccessary
COPY go.mod ./
RUN go mod download

# Copy the source code...
COPY . .

# ... and compile
RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags="-s -w -X 'main.version=$version' -X 'main.commit=$commit'" \
  -o /app/monitor /app/cmd/monitor

# === Bonus Stage: Create non-root user ===
FROM alpine:3.21 AS security_provider
RUN addgroup -S app && \
    adduser -S app -G app

# === Runtime Stage (Scratch) ===
FROM scratch
LABEL org.opencontainers.image.source=https://github.com/pgerke/freeathome

# Copy the passwd file from the security provider stage
COPY --from=security_provider /etc/passwd /etc/passwd
USER app

# Copy the compiled binary from the build stage
COPY --from=builder /app/monitor /monitor

# Set the entrypoint
ENTRYPOINT ["/monitor"]
