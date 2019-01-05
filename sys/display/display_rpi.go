// +build rpi

/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2018
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package display

import (
	"fmt"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpi "github.com/djthorpe/gopi-hw/rpi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Display struct {
	Display       uint
	PixelsPerInch string
}

type display struct {
	log      gopi.Logger
	display  uint
	handle   rpi.DX_DisplayHandle
	modeinfo rpi.DX_DisplayModeInfo
}

type NativeDisplay interface {
	Handle() rpi.DX_DisplayHandle
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open
func (config Display) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("graphics.display.Open{ display=%v }", config.Display)

	this := new(display)
	this.log = logger
	this.display = config.Display
	this.handle = rpi.DX_DISPLAY_NONE

	// Open display
	var err error
	if this.handle, err = rpi.DX_DisplayOpen(rpi.DX_DisplayId(this.display)); err != nil {
		return nil, err
	} else if this.modeinfo, err = rpi.DX_DisplayGetInfo(this.handle); err != nil {
		return nil, err
	}

	// Success
	return this, nil
}

// Close
func (this *display) Close() error {
	this.log.Debug("graphics.display.Close{ display=%v }", this.display)

	if this.handle == rpi.DX_NO_HANDLE {
		return nil
	}

	if err := rpi.DX_DisplayClose(this.handle); err != nil {
		return err
	}

	// Release resources
	this.handle = rpi.DX_NO_HANDLE
	this.modeinfo = rpi.DX_DisplayModeInfo{}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Display returns display number
func (this *display) Display() uint {
	return this.display
}

// Returns handle
func (this *display) Handle() rpi.DX_DisplayHandle {
	return this.handle
}

// Return size
func (this *display) Size() (uint32, uint32) {
	return this.modeinfo.Size.W, this.modeinfo.Size.H
}

// Return pixels-per-inch
func (this *display) PixelsPerInch() uint32 {
	// TODO
	return 0
}

// Return name of display
func (this *display) Name() string {
	return fmt.Sprint(rpi.DX_DisplayId(this.display))
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *display) String() string {
	return fmt.Sprintf("graphics.display{ id=%v (%v) info=%v }", rpi.DX_DisplayId(this.display), this.display, this.modeinfo)
}
