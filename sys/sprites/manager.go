/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package sprites

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	"github.com/djthorpe/gopi/util/errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type SpriteManager struct {
	Graphics gopi.SurfaceManager
}

type manager struct {
	log      gopi.Logger
	graphics gopi.SurfaceManager
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config SpriteManager) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<graphics.sprites>Open{ graphics=%v }", config.Graphics)

	this := new(manager)
	this.log = log
	this.graphics = config.Graphics

	return this, nil
}

func (this *manager) Close() error {
	this.log.Debug("<graphics.sprites>Close{ graphics=%v }", this.graphics)

	// Free resources
	this.graphics = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION

func (this *manager) OpenSpritesAtPath(path string, callback func(manager gopi.SpriteManager, path string, info os.FileInfo) bool) error {
	this.log.Debug2("<graphics.sprites>OpenSpritesAtPath{ path=%v }", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	errs := new(errors.CompoundError)
	fmt.Printf("XX=%v\n", errs == nil)

	errs.Add(filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
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
		// Open sprite file
		if handle, err_ := os.Open(path); err_ != nil {
			// Allow execution to continue
			errs.Add(fmt.Errorf("%v: %v", path, err_))
		} else {
			defer handle.Close()
			if _, err_ := this.OpenSprites(handle); err_ != nil {
				// Allow execution to continue
				errs.Add(fmt.Errorf("%v: %v", path, err_))
			}
		}
		// Success
		return nil
	}))

	fmt.Println(errs)

	return errs.ErrorOrSelf()
}

// Open one or more sprites from a stream and return them
func (this *manager) OpenSprites(io.Reader) ([]gopi.Sprite, error) {
	return nil, gopi.ErrNotImplemented
}

// Return loaded sprites, or a specific sprite
func (this *manager) Sprites(name string) []gopi.Sprite {
	return nil
}
