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
	if gfx := app.Graphics; gfx == nil {
		return fmt.Errorf("Missing Surfaces Manager")
	} else {
		fmt.Println(gfx)
	}

	return nil
}

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("graphics")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main))
}
