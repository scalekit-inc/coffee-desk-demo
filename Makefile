.PHONY: build build-frontend run clean install dev-frontend dev-backend docker-build

# Build the entire application
build: build-frontend
	go build -o coffee-desk-demo .

# Build the React frontend and prepare templates
build-frontend:
	cd web && npm run build
	node scripts/build-frontend.js

# Run the application
run: build
	./coffee-desk-demo

# Clean build artifacts
clean:
	rm -f coffee-desk-demo
	rm -rf internal/templates/assets
	rm -f internal/templates/*.html

# Development: run React dev server and Go backend separately
dev-frontend:
	cd web && npm run dev

dev-backend:
	go run main.go

# Install dependencies
install:
	cd web && npm install
	go mod download

# Docker build with version injection
docker-build:
	docker build \
		--build-arg VERSION=$(shell git describe --tags --always --dirty) \
		--build-arg TARGETOS=linux \
		--build-arg TARGETARCH=amd64 \
		-t coffee-desk-demo . 