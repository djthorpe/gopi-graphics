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
	_ "github.com/djthorpe/gopi-graphics/sys/surface"
	_ "github.com/djthorpe/gopi-hw/sys/hw"
	_ "github.com/djthorpe/gopi-hw/sys/metrics"
	_ "github.com/djthorpe/gopi/sys/logger"
)

type sprites struct {
	log gopi.Logger
}

func (this *sprites) OpenSpritesAtPath(path string, callback func(manager *sprites, path string, info os.FileInfo) bool) error {
	this.log.Debug2("<sprites.OpenFacesAtPath{ path=%v }", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if callback(this, path, info) == false {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}
		if info.IsDir() {
			return nil
		}
		// Open sprite
		if _, err := this.Open(path); err != nil {
			return err
		}
		// Success
		return nil
	})
	return err
}

func (this *sprites) Open(path string) (gopi.Bitmap, error) {
	this.log.Debug("<sprites>Open{ path=%v }", path)
	return nil, nil
}

func FilterFiles(manager *sprites, path string, info os.FileInfo) bool {
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
	path, _ := app.AppFlags.GetString("sprites.path")
	s := &sprites{
		log: app.Logger,
	}
	if err := s.OpenSpritesAtPath(path, FilterFiles); err != nil {
		return err
	}

	fmt.Println("Waiting for CTRL+C")
	app.WaitForSignal()
	done <- gopi.DONE
	return nil
}

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("graphics")
	config.AppFlags.FlagString("sprites.path", "", "Path for sprites")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main))
}
