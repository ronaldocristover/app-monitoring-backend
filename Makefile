.PHONY: build run test clean mocks

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

clean:
	rm -rf bin/

migrate-up:
	migrate -path ./migrations -database "postgres://app:secret@localhost:5432/app_monitoring?sslmode=disable" -up

migrate-down:
	migrate -path ./migrations -database "postgres://app:secret@localhost:5432/app_monitoring?sslmode=disable" -down

dev:
	godotenv -f .env go run ./cmd/server

# Generate mocks for handler tests
mocks:
	@echo "Generating mocks..."
	@/root/go/bin/mockery --config .mockery.yaml
