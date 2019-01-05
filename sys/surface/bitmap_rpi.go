/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package surface

import (
	"fmt"
	"image/color"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION

func (this *bitmap) Type() gopi.SurfaceType {
	return this.surface_type
}

func (this *bitmap) Size() gopi.Size {
	return this.size
}

func (this *bitmap) ClearToColorRGBA(color.RGBA) error {
	return gopi.ErrNotImplemented
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *bitmap) String() string {
	return fmt.Sprintf("<graphics.bitmap>{ id=0x%08X type=%v size=%v stride=%v }", this.handle, this.surface_type, this.size, this.stride)
}
