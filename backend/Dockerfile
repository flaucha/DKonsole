# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod file (go.sum will be generated in container)
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

ENV GOFLAGS=-mod=mod

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates && \
    addgroup -S app && adduser -S -G app -u 1000 app

WORKDIR /home/app

# Copy the binary from builder
COPY --from=builder /app/main .

# Create data directory for logo uploads
RUN mkdir -p /home/app/data && chown -R app:app /home/app

# Expose port
EXPOSE 8080

USER app

CMD ["./main"]
