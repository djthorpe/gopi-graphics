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

	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION

func (this *sprite) Name() string {
	return this.name
}

func (this *sprite) Hotspot() gopi.Point {
	return gopi.Point{float32(this.hx), float32(this.hy)}
}

func (this *sprite) Type() gopi.SurfaceFlags {
	return this.image_type
}

func (this *sprite) Size() gopi.Size {
	if this.bitmap == nil {
		return gopi.ZeroSize
	} else {
		return this.bitmap.Size()
	}
}

func (this *sprite) ClearToColor(c gopi.Color) error {
	if this.bitmap != nil {
		return this.bitmap.ClearToColor(c)
	} else {
		return gopi.ErrAppError
	}
}

func (this *sprite) FillRectToColor(c gopi.Color, o gopi.Point, s gopi.Size) error {
	if this.bitmap != nil {
		return this.bitmap.FillRectToColor(c, o, s)
	} else {
		return gopi.ErrAppError
	}
}

func (this *sprite) PaintPixel(c gopi.Color, o gopi.Point) error {
	if this.bitmap != nil {
		return this.bitmap.PaintPixel(c, o)
	} else {
		return gopi.ErrAppError
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *sprite) append(line string, linen int) error {
	// Make the array of pixel lines
	if this.pixels == nil {
		this.pixels = make([][]*pixel, 0)
	}

	// Make a pixel line
	pixel_line := make([]*pixel, 0, len(line))

	// Iterate through pixels
	for _, r := range []rune(line) {
		if p, exists := PIXEL_MAP[r]; exists == false {
			return fmt.Errorf("Invalid rune '%v' on line %v", p, linen)
		} else {
			pixel_line = append(pixel_line, &p)
		}
	}

	// Append the line
	this.pixels = append(this.pixels, pixel_line)

	// Return success
	return nil
}

func (this *sprite) create(gfx gopi.SurfaceManager) error {
	// Check incoming parameters
	if gfx == nil {
		return gopi.ErrAppError
	}

	// There should be zero or one hotspots, and determine size of the sprite
	width, height := int(0), int(0)
	hx, hy := int(0), int(0)
	for y, line := range this.pixels {
		for x, pixel := range line {
			if pixel.hotspot == true {
				if hx != 0 || hy != 0 {
					return fmt.Errorf("Only one hotspot can be defined")
				} else {
					hx, hy = x, y
				}
			}
			if x >= width {
				width = x
			}
		}
		if y >= height {
			height = y
		}
	}

	// Now create the bitmap and set up the hotspots
	if bitmap, err := gfx.CreateBitmap(this.image_type, gopi.Size{float32(width + 1), float32(height + 1)}); err != nil {
		return err
	} else if err := this.FillBitmap(bitmap); err != nil {
		gfx.DestroyBitmap(bitmap)
		return err
	} else {
		this.bitmap = bitmap
		this.hx, this.hy = hx, hy
	}

	// Success
	return nil
}

func (this *sprite) FillBitmap(bitmap gopi.Bitmap) error {

	// Clear bitmap to transparent/black
	if err := bitmap.ClearToColor(gopi.ColorWhite); err != nil {
		return err
	}
	/*
		// Stick the pixels into the bitmap
		for y, line := range this.pixels {
			for x, pixel := range line {
				// Currently we use the 'mask' bit to set transparency on the
				// pixel but it will only work with RGBA32 - for other types
				// of image we'll need to create a separate mask for the bitmap
				color := pixel.color
				if pixel.mask {
					color.A = 0.0
				}
				if err := bitmap.PaintPixel(color, gopi.Point{float32(x), float32(y)}); err != nil {
					return err
				}
			}
		}
	*/

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *sprite) String() string {
	return fmt.Sprintf("<graphics.sprite>{ name='%v' hotspot=%v bitmap=%v }", this.name, this.Hotspot(), this.bitmap)
}
