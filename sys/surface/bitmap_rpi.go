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
	"image/color"
	"unsafe"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpi "github.com/djthorpe/gopi-hw/rpi"
)

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION

func (this *bitmap) Type() gopi.SurfaceFlags {
	return this.flags.Type()
}

func (this *bitmap) Size() gopi.Size {
	return gopi.Size{float32(this.size.W), float32(this.size.H)}
}

func (this *bitmap) ClearToColorRGBA(c color.RGBA) error {
	data := make([]uint32, this.stride>>2*uint32(this.size.H))
	value := rgba_to_uint32(c)
	for i := 0; i < len(data); i++ {
		data[i] = value
	}
	ptr := uintptr(unsafe.Pointer(&data[0]))
	rect := rpi.DX_NewRect(0, 0, uint32(this.size.W), uint32(this.size.H))
	return rpi.DX_ResourceWriteData(this.handle, rpi.DX_IMAGE_TYPE_RGBA32, this.stride, ptr, rect)
}

func (this *bitmap) ClearToColor(c gopi.Color) error {
	data := make([]uint32, this.stride>>2*uint32(this.size.H))
	value := color_to_uint32(c)
	for i := 0; i < len(data); i++ {
		data[i] = value
	}
	ptr := uintptr(unsafe.Pointer(&data[0]))
	rect := rpi.DX_NewRect(0, 0, uint32(this.size.W), uint32(this.size.H))
	return rpi.DX_ResourceWriteData(this.handle, rpi.DX_IMAGE_TYPE_RGBA32, this.stride, ptr, rect)
}

func (this *bitmap) FillRectToColor(gopi.Point, gopi.Size, gopi.Color) error {
	return gopi.ErrNotImplemented
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *bitmap) String() string {
	return fmt.Sprintf("<graphics.bitmap>{ id=0x%08X flags=%v size=%v stride=%v }", this.handle, this.flags, this.size, this.stride)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func rgba_to_uint32(value color.RGBA) uint32 {
	return uint32(value.A)<<24 | uint32(value.B)<<16 | uint32(value.G)<<8 | uint32(value.R)
}

func color_to_uint32(value color.Color) uint32 {
	r, g, b, a := value.RGBA()
	return uint32(a&0xFFFF>>8)<<24 | uint32(b&0xFFFF>>8)<<16 | uint32(g&0xFFFF>>8)<<8 | uint32(r&0xFFFF>>8)
}
