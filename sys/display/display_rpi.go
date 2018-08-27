/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2017
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package display

import (
	"fmt"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpi "github.com/djthorpe/gopi-graphics/rpi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Display struct {
	Display       uint
	PixelsPerInch string
}

type display struct {
	log      gopi.Logger
	id       uint
	handle   rpi.DXDisplayHandle
	modeinfo *rpi.DXDisplayModeInfo
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open
func (config Display) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("graphics.display.Open{ id=%v }", config.Display)

	this := new(display)
	this.log = logger
	this.id = config.Display
	this.handle = rpi.DX_DISPLAY_NONE

	// Open display
	var err error
	if this.handle, err = rpi.DXDisplayOpen(this.id); err != nil {
		return nil, err
	} else if this.modeinfo, err = rpi.DXDisplayGetInfo(this.handle); err != nil {
		return nil, err
	}

	// Success
	return this, nil
}

// Close
func (this *display) Close() error {
	this.log.Debug("graphics.display.Close{ id=%v }", this.id)

	if this.handle != rpi.DX_DISPLAY_NONE {
		if err := rpi.DXDisplayClose(this.handle); err != nil {
			return err
		} else {
			this.handle = rpi.DX_DISPLAY_NONE
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Display returns display number
func (this *display) Display() uint {
	return this.id
}

// Return size
func (this *display) Size() (uint32, uint32) {
	return this.modeinfo.Size.Width, this.modeinfo.Size.Height
}

// Return pixels-per-inch
func (this *display) PixelsPerInch() uint32 {
	// TODO
	return 0
}

// Return name of display
func (this *display) Name() string {
	return fmt.Sprint(rpi.DXDisplayId(this.id))
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *display) String() string {
	return fmt.Sprintf("graphics.display{ id=%v (%v) info=%v }", rpi.DXDisplayId(this.id), this.id, this.modeinfo)
}
