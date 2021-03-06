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
	"strings"
	"sync"
	"unsafe"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	display "github.com/djthorpe/gopi-graphics/sys/display"
	egl "github.com/djthorpe/gopi-hw/egl"
	rpi "github.com/djthorpe/gopi-hw/rpi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type SurfaceManager struct {
	Display gopi.Display
}

type manager struct {
	log          gopi.Logger
	display      gopi.Display
	handle       egl.EGL_Display
	major, minor int
	surfaces     []*surface
	bitmaps      []*bitmap
	update       rpi.DX_Update
	sync.Mutex
}

type surface struct {
	log     gopi.Logger
	flags   gopi.SurfaceFlags
	opacity float32
	layer   uint16
	context egl.EGL_Context
	handle  egl.EGL_Surface
	native  *nativesurface
	bitmap  gopi.Bitmap
}

type bitmap struct {
	log             gopi.Logger
	flags           gopi.SurfaceFlags
	size            rpi.DX_Size
	handle          rpi.DX_Resource
	stride          uint32
	image_type      rpi.DX_ImageType
	bytes_per_pixel uint32
	ref             uint
	sync.Mutex
}

type bitmap_rgba struct {
	bitmap
}

type bitmap_888 struct {
	bitmap
}

type bitmap_565 struct {
	bitmap
}

type nativesurface struct {
	handle rpi.DX_Element
	size   rpi.DX_Size
	origin rpi.DX_Point
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config SurfaceManager) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<graphics.surfacemanager.Open>{ display=%v }", config.Display)

	this := new(manager)
	this.log = log

	// Check display
	this.display = config.Display
	if this.display == nil {
		return nil, gopi.ErrBadParameter
	}

	// Initialize EGL
	if handle := egl.EGL_GetDisplay(this.display.Display()); handle == nil {
		return nil, gopi.ErrBadParameter
	} else if major, minor, err := egl.EGL_Initialize(handle); err != nil {
		return nil, err
	} else {
		this.handle = handle
		this.major = major
		this.minor = minor
	}

	// Create surface array
	this.surfaces = make([]*surface, 0)
	this.bitmaps = make([]*bitmap, 0)

	return this, nil
}

func (this *manager) Close() error {
	this.log.Debug("<graphics.surfacemanager.Close>{ display=%v }", this.display)

	// Check EGL is already closed
	if this.handle == nil {
		return nil
	}

	// Free Surfaces
	if err := this.Do(func(gopi.SurfaceManager) error {
		for _, surface := range this.surfaces {
			if err := this.DestroySurface(surface); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// Free Bitmaps
	for _, bitmap := range this.bitmaps {
		if err := this.DestroyBitmap(bitmap); err != nil {
			return err
		}
	}

	// Close EGL
	if err := egl.EGL_Terminate(this.handle); err != nil {
		return err
	}

	// Free resources
	this.surfaces = nil
	this.bitmaps = nil
	this.display = nil
	this.handle = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// INTERFACE

func (this *manager) Display() gopi.Display {
	return this.display
}

func (this *manager) Name() string {
	if this.handle == nil {
		return ""
	} else {
		return fmt.Sprintf("%v %v", egl.EGL_QueryString(this.handle, egl.EGL_QUERY_VENDOR), egl.EGL_QueryString(this.handle, egl.EGL_QUERY_VERSION))
	}
}

func (this *manager) Extensions() []string {
	if this.handle == nil {
		return nil
	} else {
		return strings.Split(egl.EGL_QueryString(this.handle, egl.EGL_QUERY_EXTENSIONS), " ")
	}
}

func (this *manager) Types() []gopi.SurfaceFlags {
	if this.handle == nil {
		return nil
	}
	types := strings.Split(egl.EGL_QueryString(this.handle, egl.EGL_QUERY_CLIENT_APIS), " ")
	surface_types := make([]gopi.SurfaceFlags, 0, len(types))
	for _, t := range types {
		if t_, ok := egl.EGL_SurfaceTypeMap[t]; ok {
			surface_types = append(surface_types, t_)
		}
	}
	// always include bitmaps
	return append(surface_types, gopi.SURFACE_FLAG_BITMAP)
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *manager) String() string {
	if this.display == nil {
		return fmt.Sprintf("<graphics.surfacemanager>{ nil }")
	} else {
		return fmt.Sprintf("<graphics.surfacemanager>{ display=%v name=%v extensions=%v types=%v egl={ %v, %v }  }", this.display, this.Name(), this.Extensions(), this.Types(), this.major, this.minor)
	}
}

////////////////////////////////////////////////////////////////////////////////
// SURFACES

func (this *manager) CreateSurface(flags gopi.SurfaceFlags, opacity float32, layer uint16, origin gopi.Point, size gopi.Size) (gopi.Surface, error) {
	this.log.Debug2("<graphics.surfacemanager>CreateSurface{ flags=%v opacity=%v layer=%v origin=%v size=%v }", flags, opacity, layer, origin, size)

	// api
	api := flags.Type()

	// if Bitmap, then create a bitmap
	if api == gopi.SURFACE_FLAG_BITMAP {
		if bitmap, err := this.CreateBitmap(flags, size); err != nil {
			return nil, err
		} else if surface, err := this.CreateSurfaceWithBitmap(bitmap, flags, opacity, layer, origin, size); err != nil {
			return nil, err
		} else {
			return surface, nil
		}
	}

	// Choose r,g,b,a bits per pixel
	var r, g, b, a uint
	switch flags.Config() {
	case gopi.SURFACE_FLAG_RGB565:
		r = 5
		g = 6
		b = 5
		a = 0
	case gopi.SURFACE_FLAG_RGBA32:
		r = 8
		g = 8
		b = 8
		a = 8
	case gopi.SURFACE_FLAG_RGB888:
		r = 8
		g = 8
		b = 8
		a = 0
	default:
		return nil, gopi.ErrNotImplemented
	}

	// Create EGL context
	if api_, exists := egl.EGL_APIMap[api]; exists == false {
		return nil, gopi.ErrBadParameter
	} else if renderable_, exists := egl.EGL_RenderableMap[api]; exists == false {
		return nil, gopi.ErrBadParameter
	} else if opacity < 0.0 || opacity > 1.0 {
		return nil, gopi.ErrBadParameter
	} else if layer < gopi.SURFACE_LAYER_DEFAULT || layer > gopi.SURFACE_LAYER_MAX {
		return nil, gopi.ErrBadParameter
	} else if err := egl.EGL_BindAPI(api_); err != nil {
		return nil, err
	} else if config, err := egl.EGL_ChooseConfig(this.handle, r, g, b, a, egl.EGL_SURFACETYPE_FLAG_WINDOW, renderable_); err != nil {
		return nil, err
	} else if native_surface, err := this.CreateNativeSurface(nil, flags, opacity, layer, origin, size); err != nil {
		return nil, err
	} else if handle, err := egl.EGL_CreateSurface(this.handle, config, egl_nativewindow(native_surface)); err != nil {
		// TODO: Destroy native surface
		return nil, err
	} else if context, err := egl.EGL_CreateContext(this.handle, config, nil); err != nil {
		// TODO: Destroy native surface, window
		return nil, err
	} else if err := egl.EGL_MakeCurrent(this.handle, handle, handle, context); err != nil {
		// TODO: Destroy context, surface, window, ...
		return nil, err
	} else {
		s := &surface{
			log:     this.log,
			flags:   flags,
			opacity: opacity,
			layer:   layer,
			context: context,
			handle:  handle,
			native:  native_surface,
		}
		this.surfaces = append(this.surfaces, s)
		return s, nil
	}
}

func (this *manager) CreateSurfaceWithBitmap(bitmap gopi.Bitmap, flags gopi.SurfaceFlags, opacity float32, layer uint16, origin gopi.Point, size gopi.Size) (gopi.Surface, error) {
	flags = gopi.SURFACE_FLAG_BITMAP | bitmap.Type() | flags.Mod()
	this.log.Debug2("<graphics.surfacemanager>CreateSurfaceWithBitmap{ bitmap=%v flags=%v opacity=%v layer=%v origin=%v size=%v }", bitmap, flags, opacity, layer, origin, size)
	if opacity < 0.0 || opacity > 1.0 {
		return nil, gopi.ErrBadParameter
	} else if layer < gopi.SURFACE_LAYER_DEFAULT || layer > gopi.SURFACE_LAYER_MAX {
		return nil, gopi.ErrBadParameter
	} else if bitmap == nil {
		return nil, gopi.ErrBadParameter
	} else if size = size_from_bitmap(bitmap, size); size == gopi.ZeroSize {
		return nil, gopi.ErrBadParameter
	} else if native_surface, err := this.CreateNativeSurface(bitmap, flags, opacity, layer, origin, size); err != nil {
		return nil, err
	} else {
		// Return the surface
		s := &surface{
			log:     this.log,
			flags:   flags,
			opacity: opacity,
			layer:   layer,
			native:  native_surface,
			bitmap:  bitmap,
		}
		this.surfaces = append(this.surfaces, s)
		return s, nil
	}
}

func (this *manager) DestroySurface(s gopi.Surface) error {
	this.log.Debug2("<graphics.surfacemanager>DestroySurface{ surface=%v }", s)

	if surface_, ok := s.(*surface); ok == false {
		return gopi.ErrBadParameter
	} else {
		if surface_.handle != nil {
			if err := egl.EGL_DestroySurface(this.handle, surface_.handle); err != nil {
				return err
			} else {
				surface_.handle = nil
			}
		}
		if surface_.context != nil {
			if err := egl.EGL_DestroyContext(this.handle, surface_.context); err != nil {
				return err
			} else {
				surface_.context = nil
			}
		}
		if surface_.native != nil {
			if err := this.DestroyNativeSurface(surface_.native); err != nil {
				return err
			} else {
				surface_.native = nil
			}
		}
	}

	// Return success
	return nil
}

func (this *manager) CreateNativeSurface(b gopi.Bitmap, flags gopi.SurfaceFlags, opacity float32, layer uint16, origin gopi.Point, size gopi.Size) (*nativesurface, error) {
	this.log.Debug2("<graphics.surfacemanager>CreateNativeSurface{ bitmap=%v flags=%v opacity=%v layer=%v origin=%v size=%v }", b, flags, opacity, layer, origin, size)

	// If no update, then return out of order error
	this.Lock()
	defer this.Unlock()
	if this.update == 0 {
		return nil, gopi.ErrOutOfOrder
	}

	// Set alpha
	alpha := rpi.DX_Alpha{
		Opacity: uint32(opacity_from_float(opacity)),
	}
	if flags.Mod()&gopi.SURFACE_FLAG_ALPHA_FROM_SOURCE != 0 {
		alpha.Flags |= rpi.DX_ALPHA_FLAG_FROM_SOURCE
	} else {
		alpha.Flags |= rpi.DX_ALPHA_FLAG_FIXED_ALL_PIXELS
	}

	// Clamp, transform and protection
	clamp := rpi.DX_Clamp{}
	transform := rpi.DX_TRANSFORM_NONE
	protection := rpi.DX_PROTECTION_NONE

	// If there is a bitmap, then the source rectangle is set from that
	dest_rect := rpi.DX_NewRect(int32(origin.X), int32(origin.Y), uint32(size.W), uint32(size.H))
	src_size := rpi.DX_RectSize(dest_rect)
	dest_size := rpi.DX_RectSize(dest_rect)
	dest_origin := rpi.DX_RectOrigin(dest_rect)

	// Check size - uint16
	if src_size.W > 0xFFFF || src_size.H > 0xFFFF {
		return nil, gopi.ErrBadParameter
	}
	if dest_size.W > 0xFFFF || dest_size.H > 0xFFFF {
		return nil, gopi.ErrBadParameter
	}

	// Adjust size for source
	src_size.W = src_size.W << 16
	src_size.H = src_size.H << 16

	// Get source resource
	src_resource := rpi.DX_Resource(0)
	if b != nil {
		if bitmap_, ok := b.(*bitmap); ok == false {
			return nil, gopi.ErrBadParameter
		} else {
			src_resource = bitmap_.handle
		}
	}

	// Create the element
	if handle, err := rpi.DX_ElementAdd(this.update, rpi_dx_display(this.display), layer, dest_rect, src_resource, src_size, protection, alpha, clamp, transform); err != nil {
		return nil, err
	} else {
		return &nativesurface{handle, dest_size, dest_origin}, nil
	}
}

func (this *manager) DestroyNativeSurface(native *nativesurface) error {
	this.log.Debug2("<graphics.surfacemanager>DestroyNativeSurface{ id=0x%08X }", native.handle)

	// If no update, then return out of order error
	this.Lock()
	defer this.Unlock()
	if this.update == 0 {
		return gopi.ErrOutOfOrder
	}

	// Remove element
	if native.handle == 0 {
		return nil
	} else if err := rpi.DX_ElementRemove(this.update, native.handle); err != nil {
		return err
	} else {
		native.handle = rpi.DX_Element(0)
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// BITMAPS

func (this *manager) CreateBitmap(flags gopi.SurfaceFlags, size gopi.Size) (gopi.Bitmap, error) {
	this.log.Debug2("<graphics.surfacemanager>CreateBitmap{ flags=%v size=%v }", flags, size)

	// Check parameters
	if flags.Type() != gopi.SURFACE_FLAG_BITMAP {
		return nil, gopi.ErrBadParameter
	} else if size.W <= 0.0 || size.H <= 0.0 {
		return nil, gopi.ErrBadParameter
	}

	// Create bitmap
	b := &bitmap{
		log:   this.log,
		size:  rpi.DX_Size{uint32(size.W), uint32(size.H)},
		flags: gopi.SURFACE_FLAG_BITMAP | flags.Config(),
	}
	switch flags.Config() {
	case gopi.SURFACE_FLAG_RGBA32:
		b.image_type = rpi.DX_IMAGE_TYPE_RGBA32
		b.bytes_per_pixel = 4
	case gopi.SURFACE_FLAG_RGB888:
		b.image_type = rpi.DX_IMAGE_TYPE_RGB888
		b.bytes_per_pixel = 3
	case gopi.SURFACE_FLAG_RGB565:
		b.image_type = rpi.DX_IMAGE_TYPE_RGB565
		b.bytes_per_pixel = 2
	default:
		return nil, gopi.ErrNotImplemented
	}

	// Create resource
	if handle, err := rpi.DX_ResourceCreate(b.image_type, b.size); err != nil {
		return nil, err
	} else {
		b.handle = handle
		b.stride = rpi.DX_AlignUp(b.size.W, 16) * b.bytes_per_pixel
		this.bitmaps = append(this.bitmaps, b)
		return b, nil
	}

}

func (this *manager) DestroyBitmap(b gopi.Bitmap) error {
	this.log.Debug2("<graphics.surfacemanager>DestroyBitmap{ bitmap=%v }", b)

	if bitmap_, ok := b.(*bitmap); ok == false {
		return gopi.ErrBadParameter
	} else if bitmap_.handle != 0 {
		if err := rpi.DX_ResourceDelete(bitmap_.handle); err != nil {
			return err
		} else {
			bitmap_.handle = 0
		}
	}

	// Success
	return nil
}

func (this *manager) CreateSnapshot(flags gopi.SurfaceFlags) (gopi.Bitmap, error) {
	flags = gopi.SURFACE_FLAG_BITMAP | flags.Config() | flags.Mod()
	w, h := this.Display().Size()
	size := gopi.Size{float32(w), float32(h)}

	this.log.Debug2("<graphics.surfacemanager>CreateSnapshot{ flags=%v size=%v }", flags, size)

	if b, err := this.CreateBitmap(flags, size); err != nil {
		return nil, err
	} else if bitmap_, ok := b.(*bitmap); ok == false {
		return nil, gopi.ErrAppError
	} else if err := rpi.DX_DisplaySnapshot(rpi_dx_display(this.display), bitmap_.handle, rpi.DX_TRANSFORM_NONE); err != nil {
		return nil, err
	} else {
		return bitmap_, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func egl_nativewindow(window *nativesurface) egl.EGL_NativeWindow {
	return egl.EGL_NativeWindow(unsafe.Pointer(window))
}

func opacity_from_float(opacity float32) uint8 {
	if opacity < 0.0 {
		opacity = 0.0
	} else if opacity > 1.0 {
		opacity = 1.0
	}
	// Opacity is between 0 (fully transparent) and 255 (fully opaque)
	return uint8(opacity * float32(0xFF))
}

func rpi_dx_display(d gopi.Display) rpi.DX_DisplayHandle {
	return d.(display.NativeDisplay).Handle()
}

func size_from_bitmap(bitmap gopi.Bitmap, size gopi.Size) gopi.Size {
	if size == gopi.ZeroSize {
		return bitmap.Size()
	} else {
		return size
	}
}

////////////////////////////////////////////////////////////////////////////////
// UPDATES

func (this *manager) Do(callback gopi.SurfaceManagerCallback) error {
	if this.handle == nil {
		return gopi.ErrBadParameter
	}
	if callback == nil {
		return gopi.ErrBadParameter
	}
	if this.update != 0 {
		return gopi.ErrOutOfOrder
	}
	// TODO rpi.DX_UPDATE_PRIORITY_DEFAULT
	if update, err := rpi.DX_UpdateStart(0); err != nil {
		return err
	} else {
		this.update = update
		defer func() {
			if rpi.DX_UpdateSubmitSync(update); err != nil {
				this.log.Warn("Do: %v", err)
			}
			this.update = 0
		}()
		if err := callback(this); err != nil {
			return err
		}
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// MOVE SURFACES

func (this *manager) SetOrigin(s gopi.Surface, origin gopi.Point) error {
	this.log.Debug2("<graphics.surfacemanager>SetOrigin{ surface=%v origin=%v }", s, origin)

	// If no update, then return out of order error
	this.Lock()
	defer this.Unlock()
	if this.update == 0 {
		return gopi.ErrOutOfOrder
	}

	// Set origin
	dx_origin := rpi.DX_Point{int32(origin.X), int32(origin.Y)}

	if surface_, ok := s.(*surface); ok == false {
		return gopi.ErrBadParameter
	} else if dest_rect := rpi.DX_NewRect(dx_origin.X, dx_origin.Y, surface_.native.size.W, surface_.native.size.H); dest_rect == nil {
		return gopi.ErrBadParameter
	} else if err := rpi.DX_ElementChangeAttributes(this.update, surface_.native.handle, rpi.DX_CHANGE_FLAG_DEST_RECT, 0, 0, dest_rect, nil, 0); err != nil {
		return err
	} else {
		surface_.native.origin = dx_origin
		return nil
	}
}

func (this *manager) MoveOriginBy(s gopi.Surface, increment gopi.Point) error {
	this.log.Debug2("<graphics.surfacemanager>MoveOriginBy{ surface=%v increment=%v }", s, increment)

	// If no update, then return out of order error
	this.Lock()
	defer this.Unlock()
	if this.update == 0 {
		return gopi.ErrOutOfOrder
	}

	if surface_, ok := s.(*surface); ok == false {
		return gopi.ErrBadParameter
	} else if dest_rect := rpi.DX_NewRect(surface_.native.origin.X+int32(increment.X), surface_.native.origin.Y+int32(increment.Y), surface_.native.size.W, surface_.native.size.H); dest_rect == nil {
		return gopi.ErrBadParameter
	} else if err := rpi.DX_ElementChangeAttributes(this.update, surface_.native.handle, rpi.DX_CHANGE_FLAG_DEST_RECT, 0, 0, dest_rect, nil, 0); err != nil {
		return err
	} else {
		surface_.native.origin = rpi.DX_RectOrigin(dest_rect)
		return nil
	}
}

func (this *manager) SetLayer(s gopi.Surface, layer uint16) error {
	this.log.Debug2("<graphics.surfacemanager>SetLayer{ surface=%v layer=%v }", s, layer)

	// If no update, then return out of order error
	this.Lock()
	defer this.Unlock()
	if this.update == 0 {
		return gopi.ErrOutOfOrder
	}

	if surface_, ok := s.(*surface); ok == false {
		return gopi.ErrBadParameter
	} else if s.Layer() == gopi.SURFACE_LAYER_BACKGROUND || s.Layer() == gopi.SURFACE_LAYER_CURSOR {
		// Can't change background or cursor layers
		return gopi.ErrBadParameter
	} else if layer < gopi.SURFACE_LAYER_DEFAULT || layer > gopi.SURFACE_LAYER_MAX {
		// Invalid layer change
		return gopi.ErrBadParameter
	} else if err := rpi.DX_ElementChangeAttributes(this.update, surface_.native.handle, rpi.DX_CHANGE_FLAG_LAYER, layer, 0, nil, nil, 0); err != nil {
		return err
	} else {
		surface_.layer = layer
		return nil
	}
}

func (this *manager) SetOpacity(s gopi.Surface, opacity float32) error {
	this.log.Debug2("<graphics.surfacemanager>SetOpacity{ surface=%v opacity=%v }", s, opacity)

	// If no update, then return out of order error
	this.Lock()
	defer this.Unlock()
	if this.update == 0 {
		return gopi.ErrOutOfOrder
	}

	if surface_, ok := s.(*surface); ok == false {
		return gopi.ErrBadParameter
	} else if opacity < 0.0 || opacity > 1.0 {
		return gopi.ErrBadParameter
	} else if err := rpi.DX_ElementChangeAttributes(this.update, surface_.native.handle, rpi.DX_CHANGE_FLAG_LAYER, 0, opacity_from_float(opacity), nil, nil, 0); err != nil {
		return err
	} else {
		surface_.opacity = opacity
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// UNIMPLEMENTED

func (this *manager) SetSize(gopi.Surface, gopi.Size) error {
	return gopi.ErrNotImplemented
}

func (this *manager) SetBitmap(gopi.Bitmap) error {
	return gopi.ErrNotImplemented
}
