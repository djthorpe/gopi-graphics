/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016-2018
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"fmt"
	"os"

	// Frameworks
	"github.com/djthorpe/gopi"

	// Modules
	_ "github.com/djthorpe/gopi-graphics/sys/fonts"
	_ "github.com/djthorpe/gopi/sys/logger"
)

////////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, done chan<- struct{}) error {

	if font_manager := app.Fonts; font_manager == nil {
		return fmt.Errorf("Missing Font Manager")
	} else {
		fmt.Println(font_manager)
	}

	return nil
}

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("fonts")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main))
}
