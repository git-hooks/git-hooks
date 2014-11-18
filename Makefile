test:
	go test -v ./...

clean:
	rm -rf git-hooks_*

build:
	gox -os="linux darwin"

.PHONY: test clean build
