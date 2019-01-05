# Go parameters
GOCMD=go
GOFLAGS=-v -tags "rpi"
GOINSTALL=$(GOCMD) install $(GOFLAGS)
GOTEST=$(GOCMD) test $(GOFLAGS) 
GOCLEAN=$(GOCMD) clean

all: surface_test font_list display_list

pkg-config:
	PKG_CONFIG_PATH="/opt/vc/lib/pkgconfig"

surface_test: pkg-config
	$(GOINSTALL) ./cmd/surface_test

font_list: pkg-config
	$(GOINSTALL) ./cmd/font_list

display_list: pkg-config
	$(GOINSTALL) ./cmd/display_list

clean: 
	$(GOCLEAN)
