/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package sprites

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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

type sprite struct {
	name       string
	image_type gopi.SurfaceFlags
	size       gopi.Size
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	LINESTATE_INIT = iota
)

var (
	REGEXP_SPRITE_NAME = regexp.MustCompile("^Name:\\s*([A-Za-z0-9\\-_]+)$")
	REGEXP_SPRITE_TYPE = regexp.MustCompile("^Type:\\s*([A-Za-z0-9_]+)$")
)

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

	return errs.ErrorOrSelf()
}

// Open one or more sprites from a stream and return them
func (this *manager) OpenSprites(handle io.Reader) ([]gopi.Sprite, error) {
	// Read line by line
	scanner := bufio.NewScanner(handle)
	state := LINESTATE_INIT
	sprites := make([]gopi.Sprite, 0)
	sprite := new(sprite)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch state {
		case LINESTATE_INIT:
			// Ignore comments
			if strings.HasPrefix(line, "//") {
				continue
			} else if match := REGEXP_SPRITE_NAME.FindStringSubmatch(line); len(match) > 1 {
				sprite.name = match[1]
			} else if match := REGEXP_SPRITE_TYPE.FindStringSubmatch(line); len(match) > 1 {
				if image_type := image_type_from(match[1]); image_type == 0 {
					return nil, fmt.Errorf("Invalid image type: %v", match[1])
				} else {
					sprite.image_type = image_type
				}
			} else {
				fmt.Println(line)
			}
		default:
			return nil, gopi.ErrAppError
		}
	}
	// Add on the last sprite if not nil
	if sprite != nil {
		sprites = append(sprites, sprite)
	}

	fmt.Println(sprites)

	// Return all sprites
	return sprites, nil
}

// Return loaded sprites, or a specific sprite
func (this *manager) Sprites(name string) []gopi.Sprite {
	return nil
}

func image_type_from(value string) gopi.SurfaceFlags {
	for f := gopi.SurfaceFlags(0); f <= gopi.SurfaceFlags(gopi.SURFACE_FLAG_CONFIGMASK); f++ {
		v := "SURFACE_FLAG_BITMAP|SURFACE_FLAG_" + strings.TrimSpace(strings.ToUpper(value))
		if fmt.Sprint(f) == v {
			return f
		}
	}
	return gopi.SURFACE_FLAG_NONE
}
