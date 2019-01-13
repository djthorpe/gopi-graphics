# Go parameters
GOCMD=go
GOFLAGS=-v -tags "rpi"
GOINSTALL=$(GOCMD) install $(GOFLAGS)
GOTEST=$(GOCMD) test $(GOFLAGS) 
GOCLEAN=$(GOCMD) clean
PKG_CONFIG_PATH="/opt/vc/lib/pkgconfig"

all: surface_test font_list display_list

surface_test:
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) $(GOINSTALL) ./cmd/surface_test

sprite_test:
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) $(GOINSTALL) ./cmd/sprite_test

font_list:
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) $(GOINSTALL) ./cmd/font_list

display_list:
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) $(GOINSTALL) ./cmd/display_list

clean: 
	$(GOCLEAN)
