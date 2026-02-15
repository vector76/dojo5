.PHONY: build test dev clean frontend-build frontend-test backend-build backend-test

# Build frontend, embed assets, and compile Go binary
build: frontend-build backend-build

frontend-build:
	cd frontend && npm ci && npm run build
	rm -rf backend/internal/frontend/dist
	cp -r frontend/dist backend/internal/frontend/dist

backend-build:
	cd backend && go build -o ../dojo-crm ./cmd/server
	cp -r backend/migrations migrations

# Run both test suites
test: frontend-test backend-test

frontend-test:
	cd frontend && npm ci && npm test

backend-test:
	cd backend && go test ./...

# Run concurrent dev servers
dev:
	@echo "Starting dev servers..."
	@trap 'kill 0' EXIT; \
	(cd frontend && npm run dev) & \
	(cd backend && JWT_SECRET=dev-secret go run ./cmd/server) & \
	wait

clean:
	rm -f dojo-crm
	rm -rf migrations
	rm -rf frontend/dist
	rm -rf backend/internal/frontend/dist
