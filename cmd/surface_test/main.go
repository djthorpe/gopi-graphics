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
	_ "github.com/djthorpe/gopi-graphics/sys/surface"
	_ "github.com/djthorpe/gopi-hw/sys/hw"
	_ "github.com/djthorpe/gopi-hw/sys/metrics"
	_ "github.com/djthorpe/gopi/sys/logger"
)

////////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, done chan<- struct{}) error {

	// Create a 100x100 RGBA surface and print out the details
	size := gopi.Size{100, 100}

	if gfx := app.Graphics; gfx == nil {
		return fmt.Errorf("Missing Surfaces Manager")
	} else if surface, err := gfx.CreateSurface(gopi.SURFACE_TYPE_RGBA32, gopi.SURFACE_FLAG_NONE, 1.0, gopi.SURFACE_LAYER_DEFAULT, gopi.ZeroPoint, size); err != nil {
		return err
	} else {
		fmt.Println(surface)
	}

	return nil
}

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("display")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main))
}
