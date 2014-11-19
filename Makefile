test:
	go test -v ./...

clean:
	rm -rf git-hooks_*

build:
	gox -os="linux darwin" -output="build/{{.Dir}}_{{.OS}}_{{.Arch}}"

get:
	go get github.com/tools/godep
	godep restore ./...

.PHONY: test clean build get
