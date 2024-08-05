PKGNAME := foreman_ygg_worker

ifeq ($(origin VERSION), undefined)
	VERSION := 0.2.2
endif

go_sources := $(wildcard src/*.go)

build: $(go_sources)
	mkdir -p build
	CGO_ENABLED=0 go build -o build/foreman_worker $^

clean:
	rm -rf build

distribution-tarball:
	go mod vendor
	tar --create \
		--gzip \
		--file /tmp/$(PKGNAME)-$(VERSION).tar.gz \
		--exclude=.git \
		--exclude=.vscode \
		--exclude=.github \
		--exclude=.gitignore \
		--exclude=.copr \
		--exclude=.packit.yml \
		--transform s/^\./$(PKGNAME)-$(VERSION)/ \
		. && mv /tmp/$(PKGNAME)-$(VERSION).tar.gz .
	rm -rf ./vendor
	@echo $(PKGNAME)-$(VERSION).tar.gz

test:
	go test src/*

vet:
	go vet src/*
