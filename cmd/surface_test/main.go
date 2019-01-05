/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016-2018
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

// Outputs a table of displays - works on RPi at the moment
package main

import (
	"fmt"
	"image/color"
	"os"

	// Frameworks
	"github.com/djthorpe/gopi"

	// Modules
	_ "github.com/djthorpe/gopi-graphics/sys/display"
	_ "github.com/djthorpe/gopi-graphics/sys/surface"
	_ "github.com/djthorpe/gopi-hw/sys/hw"
	_ "github.com/djthorpe/gopi-hw/sys/metrics"
	_ "github.com/djthorpe/gopi/sys/logger"
)

////////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	if gfx := app.Graphics; gfx == nil {
		return fmt.Errorf("Missing Surfaces Manager")
	} else {
		// Create a bitmap
		if bitmap, err := gfx.CreateBitmap(gopi.SURFACE_TYPE_RGBA32, gopi.SURFACE_FLAG_NONE, gopi.Size{250, 250}); err != nil {
			return err
		} else if err := gfx.Do(func(gopi.SurfaceManager) error {
			// Clear bitmap
			bitmap.ClearToColor(color.White)
			// Create a surface and put it at { 50,50 }
			if surface, err := gfx.CreateSurfaceWithBitmap(bitmap, gopi.SURFACE_FLAG_NONE, 1.0, gopi.SURFACE_LAYER_DEFAULT, gopi.Point{50, 50}, gopi.Size{}); err != nil {
				return err
			} else {
				fmt.Println(bitmap)
				fmt.Println(surface)
				return nil
			}
		}); err != nil {
			return err
		}

		fmt.Println("Press CTRL+C to exit")
		app.WaitForSignal()
	}

	return nil
}

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("graphics")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main))
}
