run: build
	@./bin/main

build:
	@go build -ldflags "-s" -o bin/main cmd/server/main.go

test:
	go test ./...