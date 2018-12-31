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
	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register Display
	gopi.RegisterModule(gopi.Module{
		Name:     "graphics/display",
		Requires: []string{"hw"},
		Type:     gopi.MODULE_TYPE_DISPLAY,
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagUint("display", 0, "Display")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			display := Display{}
			if display_number, exists := app.AppFlags.GetUint("display"); exists {
				display.Display = display_number
			}
			return gopi.Open(display, app.Logger)
		},
	})
}
