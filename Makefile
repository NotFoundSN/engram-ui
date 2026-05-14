.PHONY: generate build run dev clean tidy

generate:
	templ generate

build: generate
	go build -o bin/engram-ui ./cmd/engram-ui

run: generate
	go run ./cmd/engram-ui

dev:
	templ generate --watch --proxy=http://localhost:7438 --cmd "go run ./cmd/engram-ui"

tidy:
	go mod tidy

clean:
	rm -rf bin/
	find . -name "*_templ.go" -delete
