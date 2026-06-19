.PHONY: build check test verify run

build:
	go build -o bin/noshot ./cmd/noshot

check:
	go vet ./...

test:
	go test ./...

verify: build check test

run:
	go run ./cmd/noshot
