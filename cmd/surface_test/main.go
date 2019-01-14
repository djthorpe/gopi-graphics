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
	_ "github.com/djthorpe/gopi-input/sys/input"
	_ "github.com/djthorpe/gopi/sys/logger"
)

////////////////////////////////////////////////////////////////////////////////

func Background(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {
	gfx := app.Graphics
	if gfx == nil {
		return fmt.Errorf("Missing Graphics Manager")
	}
	input := app.Input
	if input == nil {
		return fmt.Errorf("Missing Input Manager")
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
	fmt.Println("Press CTRL+C to exit")
	app.WaitForSignal()
	done <- gopi.DONE
	return nil
}

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("graphics", "input")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main, Background))
}
