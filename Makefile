.PHONY: build proto run test clean

# Build the binary
build:
	go build -o bin/server ./cmd/server

# Generate protobuf files
proto:
	buf generate

# Run the server
run: build
	./bin/server

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Docker build
docker-build:
	docker build -t alexandria-services .
