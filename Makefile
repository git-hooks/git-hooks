clean:
	rm -rf git-hooks_*

build:
	gox -os="linux darwin"

.PHONY: build clean
