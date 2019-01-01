# Go parameters
GOCMD=go
GOFLAGS=-tags "rpi"
GOINSTALL=$(GOCMD) install $(GOFLAGS)
GOTEST=$(GOCMD) test $(GOFLAGS) 
GOCLEAN=$(GOCMD) clean

# Freetype parameters
FT_CFLAGS=-I/usr/include/freetype2
FT_LDFLAGS=-lfreetype

# Raspberry Pi Firmware parameters
RPI_CFLAGS=-I/opt/vc/include -I/opt/vc/include/interface/vmcs_host
RPI_LDFLAGS=-L/opt/vc/lib -lbcm_host
  
all: surface_test font_list display_list

surface_test:
	CGO_CFLAGS="${RPI_CFLAGS}" CGO_LDFLAGS="${RPI_LDFLAGS}" $(GOINSTALL) ./cmd/surface_test

font_list:
	CGO_CFLAGS="${FT_CFLAGS}" CGO_LDFLAGS="${FT_LDFLAGS}" $(GOINSTALL) ./cmd/font_list

display_list:
	CGO_CFLAGS="${RPI_CFLAGS}" CGO_LDFLAGS="${RPI_LDFLAGS}" $(GOINSTALL) ./cmd/display_list

clean: 
	$(GOCLEAN)
