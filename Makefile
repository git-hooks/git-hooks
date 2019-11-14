test:
	ENV=test go test -v ./...

build:
	go build -o build/git-hooks

.PHONY: test build
