VERSION := v1.0.0

.PHONY: build clean test lint all

build:
	go build -o iurl ./cmd/iurl

clean:
	rm -f iurl iurl.exe dist/*

test:
	go test ./...

lint:
	go vet ./...

all:
	@echo "Building iurl $(VERSION) for all platforms..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -o dist/iurl-linux-amd64 ./cmd/iurl
	GOOS=linux GOARCH=arm64 go build -o dist/iurl-linux-arm64 ./cmd/iurl
	GOOS=darwin GOARCH=amd64 go build -o dist/iurl-macos-amd64 ./cmd/iurl
	GOOS=darwin GOARCH=arm64 go build -o dist/iurl-macos-arm64 ./cmd/iurl
	GOOS=windows GOARCH=amd64 go build -o dist/iurl-windows-amd64.exe ./cmd/iurl
	GOOS=freebsd GOARCH=amd64 go build -o dist/iurl-freebsd-amd64 ./cmd/iurl
	@echo "Done."
	@ls -la dist/