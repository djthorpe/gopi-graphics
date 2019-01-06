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
	"path/filepath"

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

func FilterFiles(manager gopi.SpriteManager, path string, info os.FileInfo) bool {
	if info.IsDir() {
		// Recurse into subfolders
		return true
	}
	if filepath.Ext(path) == ".sprite" {
		return true
	} else {
		return false
	}
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	if path, exists := app.AppFlags.GetString("sprites.path"); exists == false {
		return fmt.Errorf("Missing -sprites.path argument")
	} else if sprites, ok := app.ModuleInstance("graphics/sprites").(gopi.SpriteManager); ok == false {
		return fmt.Errorf("Invalid graphics/sprites component")
	} else if err := sprites.OpenSpritesAtPath(path, FilterFiles); err != nil {
		return err
	}

	fmt.Println("Waiting for CTRL+C")
	app.WaitForSignal()
	done <- gopi.DONE
	return nil
}

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("graphics/sprites")
	config.AppFlags.FlagString("sprites.path", "", "Path for sprites")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main))
}
