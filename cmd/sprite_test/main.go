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

	// Frameworks
	"github.com/djthorpe/gopi"

	// Modules
	_ "github.com/djthorpe/gopi-graphics/sys/display"
	_ "github.com/djthorpe/gopi-graphics/sys/sprites"
	_ "github.com/djthorpe/gopi-graphics/sys/surface"
	_ "github.com/djthorpe/gopi-hw/sys/hw"
	_ "github.com/djthorpe/gopi-hw/sys/metrics"
	_ "github.com/djthorpe/gopi/sys/logger"
)

func CreateCursor(gfx gopi.SurfaceManager, userInfo interface{}) error {
	cursor := userInfo.(gopi.Sprite)
	if _, err := gfx.CreateSurfaceWithBitmap(cursor, gopi.SURFACE_FLAG_NONE, 1.0, gopi.SURFACE_LAYER_DEFAULT, gopi.Point{1, 1}, gopi.ZeroSize); err != nil {
		return err
	} else {
		return nil
	}
}

func Background(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {
	if sprite := app.Sprites.Sprites("pointer_nw"); len(sprite) == 1 {
		if err := app.Graphics.Do(CreateCursor, sprite[0]); err != nil {
			return err
		}
	}

	// Now run the program
	start <- gopi.DONE

FOR_LOOP:
	for {
		select {
		case <-stop:
			break FOR_LOOP
		}
	}

	// Finished
	return nil
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {

	fmt.Println("Waiting for CTRL+C")
	app.WaitForSignal()
	done <- gopi.DONE
	return nil
}

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("sprites")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main, Background))
}
