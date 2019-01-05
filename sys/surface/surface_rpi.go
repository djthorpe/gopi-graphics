// +build rpi

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

	// Frameworks
	gopi "github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION

func (this *surface) Type() gopi.SurfaceType {
	return this.surface_type
}

func (this *surface) Size() gopi.Size {
	return this.size
}

func (this *surface) Origin() gopi.Point {
	return this.origin
}

func (this *surface) Opacity() float32 {
	return this.opacity
}

func (this *surface) Layer() uint16 {
	return this.layer
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *surface) String() string {
	return fmt.Sprintf("<graphics.surface>{ id=0x%08X type=%v size=%v origin=%v opacity=%v layer=%v }", this.native.handle, this.surface_type, this.size, this.origin, this.opacity, this.layer)
}
