# Multi-stage build for DKonsole v1.1.0
# Stage 1: Build Frontend
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy package files
COPY frontend/package*.json ./
RUN npm install

# Copy VERSION file for build-time injection
COPY VERSION ../VERSION

# Copy source code and build
COPY frontend/ .
RUN npm run build

# Stage 2: Build Backend
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app/backend

# Copy go mod file
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source code
COPY backend/ .

# Build the binary
ENV GOFLAGS=-mod=mod
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Stage 3: Final image
FROM alpine:latest

RUN apk --no-cache add ca-certificates && \
    addgroup -S app && adduser -S -G app -u 1000 app

WORKDIR /home/app

# Copy backend binary
COPY --from=backend-builder /app/backend/main .

# Copy frontend static files
COPY --from=frontend-builder /app/frontend/dist ./static

# Create data directory for logo uploads
RUN mkdir -p /home/app/data && chown -R app:app /home/app

# Expose port
EXPOSE 8080

USER app

CMD ["./main"]

