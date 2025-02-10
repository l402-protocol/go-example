.PHONY: build serve serve-gateway serve-server client client-fake clean

# Build all binaries
build:
	go build -o bin/gateway cmd/gateway/main.go
	go build -o bin/server cmd/server/main.go
	go build -o client cmd/client/main.go

# Run both servers in parallel
serve: build
	@echo "Starting backend and gateway servers..."
	@(trap 'kill 0' SIGINT; \
		go run cmd/gateway/main.go & \
		go run cmd/server/main.go & \
		wait)

# Individual server targets (for development)
serve-gateway:
	go run cmd/gateway/main.go

serve-server:
	go run cmd/server/main.go



# Clean built binaries
clean:
	rm -rf bin/
	rm -f client 