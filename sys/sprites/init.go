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
	"os"
	"path/filepath"

	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register sprites manager
	gopi.RegisterModule(gopi.Module{
		Name:     "graphics/sprites",
		Type:     gopi.MODULE_TYPE_SPRITES,
		Requires: []string{"graphics"},
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagString("sprites.path", "", "Path for sprite files")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(SpriteManager{
				Graphics: app.Graphics,
			}, app.Logger)
		},
		Run: func(app *gopi.AppInstance, sprites gopi.Driver) error {
			if path, exists := app.AppFlags.GetString("sprites.path"); exists && len(path) > 0 {
				if err := sprites.(gopi.SpriteManager).OpenSpritesAtPath(path, FilterFiles); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

////////////////////////////////////////////////////////////////////////////////
// FILTER SPRITE FILES BY EXTENSION

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
