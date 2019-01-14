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
	return gopi.ZeroPoint
}

func (this *sprite) Type() gopi.SurfaceFlags {
	return this.image_type
}

func (this *sprite) Size() gopi.Size {
	return this.size
}

func (this *sprite) ClearToColor(gopi.Color) error {
	return gopi.ErrNotImplemented
}

func (this *sprite) FillRectToColor(gopi.Point, gopi.Size, gopi.Color) error {
	return gopi.ErrNotImplemented
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *sprite) String() string {
	return fmt.Sprintf("<graphics.sprite>{ name='%v' type=%v size=%v }", this.name, this.image_type, this.size)
}
