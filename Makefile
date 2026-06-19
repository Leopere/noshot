.PHONY: build check test verify run package

build:
	go build -o bin/noshot ./cmd/noshot

check:
	go vet ./...

test:
	go test ./...

verify: build check test

run:
	go run ./cmd/noshot

package:
	./scripts/package-macos.sh
