// +build rpi

/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package sprites

import (
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register sprites manager
	gopi.RegisterModule(gopi.Module{
		Name:     "graphics/sprites",
		Type:     gopi.MODULE_TYPE_OTHER,
		Requires: []string{"graphics"},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(SpriteManager{
				Graphics: app.Graphics,
			}, app.Logger)
		},
	})
}
