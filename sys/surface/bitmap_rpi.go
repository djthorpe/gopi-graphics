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
	"unsafe"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpi "github.com/djthorpe/gopi-hw/rpi"
)

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION

func (this *bitmap) Type() gopi.SurfaceFlags {
	return this.flags.Config()
}

func (this *bitmap) Size() gopi.Size {
	return gopi.Size{float32(this.size.W), float32(this.size.H)}
}

func (this *bitmap) ClearToColor(c gopi.Color) error {
	this.log.Debug2("<graphics.surfacemanager>ClearToColor{ color=%v }", c)

	// Create a strip of data
	data := make([]byte, 0, this.stride)
	src := color_to_bytes(c, this.image_type)
	for i := uint32(0); i < this.size.W; i++ {
		data = append(data, src...)
	}
	// Set the pointer to the strip and move y forward and ptr back for each strip
	ptr := uintptr(unsafe.Pointer(&data[0]))
	rect := rpi.DX_NewRect(0, 0, uint32(this.size.W), 1)
	for y := uint32(0); y < this.size.H; y++ {
		rpi.DX_RectSet(rect, 0, int32(y), uint32(this.size.W), 1)
		if err := rpi.DX_ResourceWriteData(this.handle, this.image_type, this.stride, ptr, rect); err != nil {
			return err
		}
		// Offset pointer backwards
		ptr -= uintptr(this.stride)
	}
	// Return success
	return nil
}

func (this *bitmap) FillRectToColor(color gopi.Color, origin gopi.Point, size gopi.Size) error {
	this.log.Debug2("<graphics.surfacemanager>FillRectToColor{ color=%v origin=%v size=%v }", color, origin, size)

	// Calculate the intersection between the the rectangle and the surface frame
	// If width or height is zero there is no intersection to return error
	frame := rpi.DX_NewRect(0, 0, uint32(this.size.W), uint32(this.size.H))
	rect := rpi.DX_NewRect(int32(origin.X), int32(origin.Y), uint32(size.W), uint32(size.H))
	if intersection := rpi.DX_RectIntersection(frame, rect); intersection == nil {
		return nil
	} else if origin, size := rpi.DX_RectOrigin(intersection), rpi.DX_RectSize(intersection); origin.X == 0 && origin.Y == 0 && size.W == this.size.W && size.H == this.size.H {
		// Intersection is the whole image, so use 'ClearToColor'
		return this.ClearToColor(color)
	} else if size.W > 0 && size.H > 0 {
		// Create a strip of data
		data := make([]byte, 0, size.W*this.bytes_per_pixel)
		src := color_to_bytes(color, this.image_type)
		for i := uint32(0); i < size.W; i++ {
			data = append(data, src...)
		}
		// Set the pointer to the strip and move y forward and ptr back for each strip
		ptr := uintptr(unsafe.Pointer(&data[0]))
		rect := rpi.DX_NewRect(0, 0, uint32(size.W), 1)
		for y := uint32(0); y < size.H; y++ {
			rpi.DX_RectSet(rect, 0, int32(y), uint32(size.W), 1)
			if err := rpi.DX_ResourceWriteData(this.handle, this.image_type, uint32(len(data)), ptr, rect); err != nil {
				return err
			}
			// Offset pointer backwards
			ptr -= uintptr(len(data))
		}
	}
	return nil
}

func (this *bitmap) PaintPixel(color gopi.Color, origin gopi.Point) error {
	this.log.Debug2("<graphics.surfacemanager>PaintPixel{ color=%v origin=%v }", color, origin)
	return gopi.ErrNotImplemented
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *bitmap) String() string {
	return fmt.Sprintf("<graphics.bitmap>{ id=0x%08X type=%v size=%v stride=%v }", this.handle, this.flags.ConfigString(), this.size, this.stride)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func color_to_bytes(c gopi.Color, t rpi.DX_ImageType) []byte {
	// Returns color 0000 <= v <= FFFF
	r, g, b, a := c.RGBA()
	// Convert to []byte
	switch t {
	case rpi.DX_IMAGE_TYPE_RGB888:
		return []byte{byte(r >> 8), byte(g >> 8), byte(b >> 8)}
	case rpi.DX_IMAGE_TYPE_RGB565:
		r := uint16(r>>(8+3)) << (5 + 6)
		g := uint16(g>>(8+2)) << 5
		b := uint16(b >> (8 + 3))
		v := r | g | b
		return []byte{byte(v), byte(v >> 8)}
	case rpi.DX_IMAGE_TYPE_RGBA32:
		return []byte{byte(r >> 8), byte(g >> 8), byte(b >> 8), byte(a >> 8)}
	default:
		return nil
	}
}
