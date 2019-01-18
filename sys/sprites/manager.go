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
	sprites  map[string]gopi.Sprite
}

type sprite struct {
	name       string
	image_type gopi.SurfaceFlags
	pixels     [][]*pixel
	hx, hy     int
	bitmap     gopi.Bitmap
}

type pixel struct {
	color   gopi.Color
	hotspot bool
	mask    bool
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	LINESTATE_INIT = iota
	LINESTATE_DATA
)

var (
	REGEXP_SPRITE_NAME = regexp.MustCompile("^Name:\\s*([A-Za-z0-9\\-_]+)$")
	REGEXP_SPRITE_TYPE = regexp.MustCompile("^Type:\\s*([A-Za-z0-9_]+)$")
	REGEXP_SPRITE_DATA = regexp.MustCompile("^\\s*([A-Za-z0-9\\.\\-]+)\\s*$")
)

var (
	PIXEL_MAP = map[rune]pixel{
		'b': pixel{gopi.ColorBlack, false, false},
		'B': pixel{gopi.ColorBlack, true, false},
		'w': pixel{gopi.ColorWhite, false, false},
		'W': pixel{gopi.ColorWhite, true, false},
		'.': pixel{gopi.ColorBlack, false, true},
	}
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config SpriteManager) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<graphics.sprites>Open{ graphics=%v }", config.Graphics)

	// Check parameters
	if config.Graphics == nil {
		log.Warn("graphics.sprites: Missing surface manager")
		return nil, gopi.ErrBadParameter
	}

	this := new(manager)
	this.log = log
	this.graphics = config.Graphics
	this.sprites = make(map[string]gopi.Sprite, 0)

	return this, nil
}

func (this *manager) Close() error {
	this.log.Debug("<graphics.sprites>Close{ graphics=%v }", this.graphics)

	// Free resources
	this.graphics = nil
	this.sprites = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *manager) String() string {
	parts := make([]string, 0, len(this.sprites))
	for _, sprite := range this.sprites {
		parts = append(parts, fmt.Sprint(sprite))
	}
	return fmt.Sprintf("<graphics.sprites>{ %v }", strings.Join(parts, ","))
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
	sprite_ := new(sprite)
	linen := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		linen += 1
		switch state {
		case LINESTATE_INIT:
			// Ignore comments, spaces
			if strings.HasPrefix(line, "//") || len(line) == 0 {
				continue
			} else if match := REGEXP_SPRITE_NAME.FindStringSubmatch(line); len(match) > 1 {
				sprite_.name = match[1]
			} else if match := REGEXP_SPRITE_TYPE.FindStringSubmatch(line); len(match) > 1 {
				if image_type := image_type_from(match[1]); image_type == 0 {
					return nil, fmt.Errorf("Invalid image type: %v", match[1])
				} else {
					sprite_.image_type = image_type
				}
			} else if match := REGEXP_SPRITE_DATA.FindStringSubmatch(line); len(match) > 1 {
				if err := sprite_.append(line, linen); err != nil {
					return nil, err
				}
				state = LINESTATE_DATA
			} else {
				return nil, fmt.Errorf("Syntax error on line %v", linen)
			}
		case LINESTATE_DATA:
			if match := REGEXP_SPRITE_DATA.FindStringSubmatch(line); len(match) > 1 {
				if err := sprite_.append(line, linen); err != nil {
					return nil, err
				}
			} else if strings.HasPrefix(line, "//") || len(line) == 0 {
				// Eject and next sprite
				sprites = append(sprites, sprite_)
				sprite_ = new(sprite)
			} else {
				return nil, fmt.Errorf("Syntax error on line %v", linen)
			}
		default:
			return nil, gopi.ErrAppError
		}
	}

	// Add on the last sprite if not nil
	if sprite_ != nil && sprite_.Size() != gopi.ZeroSize {
		sprites = append(sprites, sprite_)
		sprite_ = nil
	}

	// For each sprite, create the bitmaps
	for _, sprite_ := range sprites {
		if err := sprite_.(*sprite).create(this.graphics); err != nil {
			return nil, err
		} else if name := sprite_.Name(); len(name) == 0 {
			return nil, fmt.Errorf("Sprite with missing name")
		} else if _, exists := this.sprites[name]; exists {
			return nil, fmt.Errorf("Duplicate sprite named '%v'", name)
		} else {
			this.sprites[name] = sprite_
		}
	}

	// Return all sprites
	return sprites, nil
}

// Return loaded sprites, or a specific sprite
func (this *manager) Sprites(name string) []gopi.Sprite {
	if name != "" {
		if sprite, exists := this.sprites[name]; exists {
			return []gopi.Sprite{sprite}
		} else {
			return nil
		}
	}
	sprites := make([]gopi.Sprite, 0, len(this.sprites))
	for _, sprite := range this.sprites {
		sprites = append(sprites, sprite)
	}
	return sprites
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
