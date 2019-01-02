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

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	egl "github.com/djthorpe/gopi-hw/egl"
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
	surface      egl.EGL_Surface
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

	// Free Surfaces
	for _, surface := range this.surfaces {
		if err := this.DestroySurface(surface); err != nil {
			return err
		}
	}

	// TODO: Free Bitmaps

	// Close EGL
	if err := egl.EGL_Terminate(this.handle); err != nil {
		return err
	}

	// Free resources
	this.surfaces = nil
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
	} else if context, err := egl.EGL_CreateContext(this.handle, config, nil); err != nil {
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
		}
		this.surfaces = append(this.surfaces, s)
		return s, nil
	}
}

func (this *manager) CreateSurfaceWithBitmap(bitmap gopi.Bitmap, flags gopi.SurfaceFlags, opacity float32, layer uint16, origin gopi.Point, size gopi.Size) (gopi.Surface, error) {
	this.log.Debug2("<graphics.surfacemanager>CreateSurfaceWithBitmap{ bitmap=%v flags=%v opacity=%v layer=%v origin=%v size=%v }", bitmap, flags, opacity, layer, origin, size)
	return nil, gopi.ErrNotImplemented
}

func (this *manager) DestroySurface(s gopi.Surface) error {
	this.log.Debug2("<graphics.surfacemanager>DestroySurface{ surface=%v }", s)

	if surface_, ok := s.(*surface); ok == false {
		return gopi.ErrBadParameter
	} else {
		if surface_.surface != nil {
			if err := egl.EGL_DestroySurface(this.handle, surface_.surface); err != nil {
				return err
			} else {
				surface_.surface = nil
			}
		}

		if surface_.context != nil {
			if err := egl.EGL_DestroyContext(this.handle, surface_.context); err != nil {
				return err
			} else {
				surface_.context = nil
			}
		}
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// BITMAPS

func (this *manager) CreateBitmap(api gopi.SurfaceType, size gopi.Size) (gopi.Bitmap, error) {
	return nil, gopi.ErrNotImplemented
}

func (this *manager) DestroyBitmap(bitmap gopi.Bitmap) error {
	return gopi.ErrNotImplemented
}

/*
////////////////////////////////////////////////////////////////////////////////
// DO

func (this *manager) Do(callback gopi.SurfaceManagerCallback) error {
	// check parameters
	if this.handle == rpi.EGLDisplay(rpi.EGL_NO_DISPLAY) {
		return gopi.ErrBadParameter
	}

	// create update
	if err := this.doUpdateStart(); err != nil {
		return err
	}

	// callback
	cb_err := callback(this)

	// end update
	if err := this.doUpdateEnd(); err != nil {
		this.log.Error("doUpdateEnd: %v", err)
	}

	// return callback error
	return cb_err
}

func (this *manager) doUpdateStart() error {
	this.Lock()
	defer this.Unlock()
	if this.update != rpi.DXUpdateHandle(rpi.DX_NO_UPDATE) {
		return gopi.ErrOutOfOrder
	}
	if update, err := rpi.DXUpdateStart(rpi.DX_UPDATE_PRIORITY_DEFAULT); err != rpi.DX_SUCCESS {
		return os.NewSyscallError("DXUpdateStart", err)
	} else {
		this.update = update
		return nil
	}
}

func (this *manager) doUpdateEnd() error {
	this.Lock()
	defer this.Unlock()
	if this.update == rpi.DXUpdateHandle(rpi.DX_NO_UPDATE) {
		return gopi.ErrOutOfOrder
	}
	if err := rpi.DXUpdateSubmitSync(this.update); err != rpi.DX_SUCCESS {
		return os.NewSyscallError("DXUpdateSubmitSync", err)
	} else {
		this.update = rpi.DXUpdateHandle(rpi.DX_NO_UPDATE)
		return nil
	}
}

// SetLayer changes a surface layer (except if it's a background or cursor). Currently
// the flags argument is ignored
func (this *manager) SetLayer(surface gopi.Surface, flags gopi.SurfaceFlags, layer uint16) error {
	return gopi.ErrNotImplemented
}

// SetOrigin moves the surface. Currently the flags argument is ignored
func (this *manager) SetOrigin(surface gopi.Surface, flags gopi.SurfaceFlags, origin gopi.Point) error {
	return gopi.ErrNotImplemented
}

func (this *manager) SetOpacity(surface gopi.Surface, flags gopi.SurfaceFlags, opacity float32) error {
	return gopi.ErrNotImplemented
}

*/
