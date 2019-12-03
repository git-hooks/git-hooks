PLATFORMS := darwin/386 darwin/amd64 linux/386 linux/amd64 linux/arm 

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

release: $(PLATFORMS)

$(PLATFORMS):
	GOOS=$(os) GOARCH=$(arch) go build -o 'build/git-hooks_$(os)-$(arch)'

test:
	ENV=test go test -v ./...

build:
	go build -o build/git-hooks

.PHONY: test build release $(PLATFORMS) clean

clean:
	rm -rf build/*
