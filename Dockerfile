# Multi-stage build for DKonsole
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

# Verify build output
RUN ls -la /app/frontend/dist/ || (echo "ERROR: Frontend build failed - dist directory not found" && exit 1)

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

# Create directories with correct permissions
RUN mkdir -p /home/app/data /home/app/static && chown -R app:app /home/app

# Copy frontend static files from frontend builder
# Using --chown to set ownership directly during copy
COPY --from=frontend-builder --chown=app:app /app/frontend/dist /home/app/static/

# Verify static files were copied (as root before switching user)
RUN ls -la /home/app/static/ || (echo "ERROR: Static files not copied" && exit 1)
RUN test -f /home/app/static/index.html || (echo "ERROR: index.html not found in static directory" && exit 1)

# Ensure proper permissions
RUN chown -R app:app /home/app

# Expose port
EXPOSE 8080

USER app

CMD ["./main"]
