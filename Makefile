test:
	go test -v ./...

clean:
	rm -rf build/*

build:
	gox -os="linux darwin" -output="build/{{.Dir}}_{{.OS}}_{{.Arch}}"

get:
	go get github.com/tools/godep
	godep restore ./...

.PHONY: test clean build get
