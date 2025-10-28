# Stage 1: Build the frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app
COPY web/package.json web/package-lock.json ./
RUN npm install
COPY web/ ./
RUN npm run build

# Stage 2: Build the backend
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app

# Install Node.js to run the build script
RUN apk add --no-cache nodejs npm

# Copy go module files
COPY go.mod go.sum ./

# Download Go dependencies
RUN go mod download

# Copy application source
COPY . .

# Copy frontend build artifacts
COPY --from=frontend-builder /app/dist ./web/dist

# Generate templates from frontend build
RUN node scripts/build-frontend.js

# Build the Go binary with optimization
RUN CGO_ENABLED=0 go build \
    -ldflags="-w -s" \
    -o /coffee-desk-demo .

# Stage 3: Create the final, minimal image using distroless
FROM gcr.io/distroless/static-debian12:debug

# Copy the static binary from the backend-builder stage
COPY --from=backend-builder /coffee-desk-demo /app/coffee-desk-demo

WORKDIR /app

EXPOSE 8080

# Run the binary
ENTRYPOINT ["/app/coffee-desk-demo"] 