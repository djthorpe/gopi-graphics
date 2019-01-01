// +build rpi

/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2017
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package surface

import (
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register surface manager
	gopi.RegisterModule(gopi.Module{
		Name:     "graphics/surfaces",
		Type:     gopi.MODULE_TYPE_GRAPHICS,
		Requires: []string{"display"},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(SurfaceManager{
				Display: app.Display,
			}, app.Logger)
		},
	})
}
