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
	"os"
	"time"

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

func Background(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {
	gfx := app.Graphics
	if gfx == nil {
		return fmt.Errorf("Missing Surfaces Manager")
	}

	var surface1, surface2, surface3 gopi.Surface

	// Create a 2x2 bitmap and place on the screen
	if bitmap1, err := gfx.CreateBitmap(gopi.SURFACE_FLAG_BITMAP, gopi.Size{2, 2}); err != nil {
		return err
	} else if bitmap2, err := gfx.CreateBitmap(gopi.SURFACE_FLAG_BITMAP|gopi.SURFACE_FLAG_RGB888, gopi.Size{10, 10}); err != nil {
		return err
	} else if bitmap3, err := gfx.CreateSnapshot(gopi.SURFACE_FLAG_RGB565); err != nil {
		return err
	} else if err := gfx.Do(func(gopi.SurfaceManager) error {
		// Clear bitmaps to partly translucent color
		bitmap1.ClearToColor(gopi.Color{1.0, 0, 0, 0.8})
		bitmap2.ClearToColor(gopi.Color{0, 0, 1.0, 0.8})
		bitmap2.FillRectToColor(gopi.ZeroPoint, gopi.Size{10, 1}, gopi.ColorPurple)

		// Create surfaces
		if s, err := gfx.CreateSurfaceWithBitmap(bitmap1, gopi.SURFACE_FLAG_ALPHA_FROM_SOURCE, 1.0, gopi.SURFACE_LAYER_DEFAULT, gopi.Point{50, 50}, gopi.Size{250, 250}); err != nil {
			return err
		} else {
			surface1 = s
		}
		if s, err := gfx.CreateSurfaceWithBitmap(bitmap2, 0, 1.0, gopi.SURFACE_LAYER_DEFAULT, gopi.Point{250, 250}, gopi.Size{250, 250}); err != nil {
			return err
		} else {
			surface2 = s
		}
		if s, err := gfx.CreateSurfaceWithBitmap(bitmap3, 0, 1.0, gopi.SURFACE_LAYER_DEFAULT, gopi.Point{150, 150}, bitmap3.Size()); err != nil {
			return err
		} else {
			surface3 = s
		}
		return nil
	}); err != nil {
		return err
	}

	// Now run the program
	start <- gopi.DONE

	// Move bitmap once per second
	timer := time.NewTicker(time.Millisecond * 1)
FOR_LOOP:
	for {
		select {
		case <-timer.C:
			gfx.Do(func(gfx gopi.SurfaceManager) error {
				gfx.MoveOriginBy(surface1, gopi.Point{1, 1})
				gfx.MoveOriginBy(surface2, gopi.Point{-1, -1})
				gfx.MoveOriginBy(surface3, gopi.Point{-2, 2})
				return nil
			})
		case <-stop:
			break FOR_LOOP
		}
	}

	// Finished
	return nil
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	fmt.Println("Press CTRL+C to exit")
	app.WaitForSignal()
	done <- gopi.DONE
	return nil
}

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("graphics")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main, Background))
}
