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
	log          gopi.Logger
	surface_type gopi.SurfaceType
	size         gopi.Size
	origin       gopi.Point
	opacity      float32
	layer        uint16
	context      egl.EGL_Context
	handle       egl.EGL_Surface
	native       *nativesurface
}

type bitmap struct {
	log          gopi.Logger
	surface_type gopi.SurfaceType
	size         gopi.Size
	handle       rpi.DX_Resource
	stride       uint32
}

type nativesurface struct {
	handle rpi.DX_Element
	width  int
	height int
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

	return this, nil
}

func (this *manager) Close() error {
	this.log.Debug("<graphics.surfacemanager.Close>{ display=%v }", this.display)

	// Check EGL is already closed
	if this.handle == nil {
		return nil
	}

	// TODO: Start Update

	// Free Surfaces
	for _, surface := range this.surfaces {
		if err := this.DestroySurface(surface); err != nil {
			return err
		}
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

func (this *manager) Types() []gopi.SurfaceType {
	if this.handle == nil {
		return nil
	}
	types := strings.Split(egl.EGL_QueryString(this.handle, egl.EGL_QUERY_CLIENT_APIS), " ")
	surface_types := make([]gopi.SurfaceType, 0, len(types))
	for _, t := range types {
		if t_, ok := egl.EGL_SurfaceTypeMap[t]; ok {
			surface_types = append(surface_types, t_)
		}
	}
	// always include RGBA32
	return append(surface_types, gopi.SURFACE_TYPE_RGBA32)
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

func (this *manager) CreateSurface(api gopi.SurfaceType, flags gopi.SurfaceFlags, opacity float32, layer uint16, origin gopi.Point, size gopi.Size) (gopi.Surface, error) {
	this.log.Debug2("<graphics.surfacemanager>CreateSurface{ api=%v flags=%v opacity=%v layer=%v origin=%v size=%v }", api, flags, opacity, layer, origin, size)

	// Create EGL context with 8 bits per pixel, 8 bits for ALpha
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
	} else if config, err := egl.EGL_ChooseConfig(this.handle, 8, 8, egl.EGL_SURFACETYPE_FLAG_WINDOW, renderable_); err != nil {
		return nil, err
	} else if native_surface, err := this.CreateNativeSurface(nil, flags, opacity, layer, origin, size); err != nil {
		return nil, err
	} else if handle, err := egl.EGL_CreateSurface(this.handle, config, egl_nativewindow(native_surface)); err != nil {
		// TODO: Create native surface
		return nil, err
	} else if context, err := egl.EGL_CreateContext(this.handle, config, nil); err != nil {
		// TODO: Create native surface, window
		return nil, err
	} else if err := egl.EGL_MakeCurrent(this.handle, handle, handle, context); err != nil {
		// TODO destroy context, surface, window, ...
		return nil, err
	} else {
		s := &surface{
			log:          this.log,
			surface_type: api,
			size:         size,
			origin:       origin,
			opacity:      opacity,
			layer:        layer,
			context:      context,
			handle:       handle,
			native:       native_surface,
		}
		this.surfaces = append(this.surfaces, s)
		return s, nil
	}
}

func (this *manager) CreateSurfaceWithBitmap(bitmap gopi.Bitmap, flags gopi.SurfaceFlags, opacity float32, layer uint16, origin gopi.Point, size gopi.Size) (gopi.Surface, error) {
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
		s := &surface{
			log:          this.log,
			surface_type: bitmap.Type(),
			size:         size,
			origin:       origin,
			opacity:      opacity,
			layer:        layer,
			native:       native_surface,
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
			if err := this.DestroyNativeSurface(s); err != nil {
				return err
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
		Opacity: opacity_from_float(opacity),
	}
	if flags&gopi.SURFACE_FLAG_ALPHA_FROM_SOURCE != 0 {
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

	fmt.Printf("dest_rect=%v src_size=%v src_resource=%X\n", dest_rect, src_size, src_resource)

	// Create the element
	if handle, err := rpi.DX_ElementAdd(this.update, rpi_dx_display(this.display), layer, dest_rect, src_resource, src_size, protection, alpha, clamp, transform); err != nil {
		return nil, err
	} else {
		return &nativesurface{handle, int(dest_size.W), int(dest_size.H)}, nil
	}
}

func (this *manager) DestroyNativeSurface(s gopi.Surface) error {
	this.log.Debug2("<graphics.surfacemanager>DestroyNativeSurface{ surface=%v }", s)

	// If no update, then return out of order error
	this.Lock()
	defer this.Unlock()
	if this.update == 0 {
		return gopi.ErrOutOfOrder
	}

	// If no native element, return
	if surface_, ok := s.(*surface); ok == false {
		return gopi.ErrBadParameter
	} else if surface_.native == nil {
		return nil
	} else if err := rpi.DX_ElementRemove(this.update, surface_.native.handle); err != nil {
		return err
	} else {
		surface_.native = nil
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// BITMAPS

func (this *manager) CreateBitmap(api gopi.SurfaceType, flags gopi.SurfaceFlags, size gopi.Size) (gopi.Bitmap, error) {
	this.log.Debug2("<graphics.surfacemanager>CreateBitmap{ api=%v flags=%v size=%v }", api, flags, size)
	if api != gopi.SURFACE_TYPE_RGBA32 {
		return nil, gopi.ErrBadParameter
	}
	if size.W <= 0.0 || size.H <= 0.0 {
		return nil, gopi.ErrBadParameter
	}
	if handle, err := rpi.DX_ResourceCreate(rpi.DX_IMAGE_TYPE_RGBA32, rpi.DX_Size{uint32(size.W), uint32(size.H)}); err != nil {
		return nil, err
	} else {
		// Alignment on (4 x uint32) boundaries
		b := &bitmap{
			log:          this.log,
			surface_type: api,
			size:         size,
			handle:       handle,
			stride:       rpi.DX_AlignUp(uint32(size.W), 16) * 4,
		}
		this.bitmaps = append(this.bitmaps, b)
		return b, nil
	}
}

func (this *manager) DestroyBitmap(b gopi.Bitmap) error {
	this.log.Debug2("<graphics.surfacemanager>DestroyBitmap{ bitmap=%v }", b)

	if bitmap_, ok := b.(*bitmap); ok == false {
		return gopi.ErrBadParameter
	} else {
		if bitmap_.handle != 0 {
			if err := rpi.DX_ResourceDelete(bitmap_.handle); err != nil {
				return err
			} else {
				bitmap_.handle = 0
			}
		}
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func egl_nativewindow(window *nativesurface) egl.EGL_NativeWindow {
	return egl.EGL_NativeWindow(unsafe.Pointer(window))
}

func opacity_from_float(opacity float32) uint32 {
	if opacity < 0.0 {
		opacity = 0.0
	} else if opacity > 1.0 {
		opacity = 1.0
	}
	return uint32(opacity * float32(0xFF))
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
// UNIMPLEMENTED

func (this *manager) SetOrigin(gopi.Surface, gopi.Point) error {
	return gopi.ErrNotImplemented
}
func (this *manager) MoveOriginBy(gopi.Surface, gopi.Point) error {
	return gopi.ErrNotImplemented
}
func (this *manager) SetSize(gopi.Surface, gopi.Size) error {
	return gopi.ErrNotImplemented
}
func (this *manager) SetLayer(gopi.Surface, uint16) error {
	return gopi.ErrNotImplemented
}
func (this *manager) SetOpacity(gopi.Surface, float32) error {
	return gopi.ErrNotImplemented
}
func (this *manager) SetBitmap(gopi.Bitmap) error {
	return gopi.ErrNotImplemented
}
