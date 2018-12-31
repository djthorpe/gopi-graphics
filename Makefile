# Go parameters
GOCMD=go
GOFLAGS=-tags "rpi"
GOINSTALL=$(GOCMD) install $(GOFLAGS)
GOTEST=$(GOCMD) test $(GOFLAGS) 
GOCLEAN=$(GOCMD) clean
    
all: test install

install:
	$(GOINSTALL) ./cmd/display_list
	$(GOINSTALL) ./cmd/font_list
	$(GOINSTALL) ./cmd/surface_test

test: 
	$(GOTEST) -v ./...

clean: 
	$(GOCLEAN)
