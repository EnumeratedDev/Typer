# Installation paths
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
SYSCONFDIR := $(PREFIX)/etc

# Compilers and tools
GO ?= $(shell which go)

build:
	mkdir -p build
	cd src/; $(GO) build -ldflags "-w -X 'main.sysconfdir=$(SYSCONFDIR)'" -o ../build/typer

install: build/typer
	# Create directories
	install -dm755 $(DESTDIR)$(BINDIR)
	install -dm755 $(DESTDIR)$(SYSCONFDIR)
	# Install files
	install -Dm755 build/typer $(DESTDIR)$(BINDIR)/typer
	cp -r config -T $(DESTDIR)$(SYSCONFDIR)/typer

uninstall:
	rm $(DESTDIR)$(BINDIR)/typer

clean:
	rm -r build/

.PHONY: build
