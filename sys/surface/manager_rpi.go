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
	handle       egl.Display
	major, minor int
	sync.Mutex
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
	egl_display := egl.EGL_GetDisplay(this.display.Display())
	if major, minor, err := egl.EGL_Initialize(egl_display); err != nil {
		return nil, err
	} else {
		this.handle = handle
		this.major = major
		this.minor = minor
	}

	return this, nil
}

func (this *manager) Close() error {
	this.log.Debug("<graphics.surfacemanager.Close>{ display=%v }", this.display)

	// Check display is already closed
	if this.display == nil {
		return nil
	}

	// TODO: Free Surfaces and Bitmaps
	// Close EGL
	egl_display := egl.EGL_GetDisplay(this.display.Display())
	if err := egl.EGL_Terminate(egl_display); err != nil {
		return err
	}

	// Blank out
	this.display = nil
	this.handle = nil

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *manager) String() string {
	if this.display == nil {
		return fmt.Sprintf("<graphics.surfacemanager>{ nil }")
	} else {
		return fmt.Sprintf("<graphics.surfacemanager>{ display=%v egl={ %v, %v } }", this.display, this.major, this.minor)
	}
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

////////////////////////////////////////////////////////////////////////////////
// SURFACE

func (this *manager) CreateSurface(api gopi.SurfaceType, flags gopi.SurfaceFlags, opacity float32, layer uint16, origin gopi.Point, size gopi.Size) (gopi.Surface, error) {
	return nil, gopi.ErrNotImplemented
}

func (this *manager) CreateSurfaceWithBitmap(bitmap gopi.Bitmap, flags gopi.SurfaceFlags, opacity float32, layer uint16, origin gopi.Point, size gopi.Size) (gopi.Surface, error) {
	return nil, gopi.ErrNotImplemented
}

func (this *manager) DestroySurface(surface gopi.Surface) error {
	return gopi.ErrNotImplemented
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

////////////////////////////////////////////////////////////////////////////////
// BITMAP

func (this *manager) CreateBitmap(api gopi.SurfaceType, size gopi.Size) (gopi.Bitmap, error) {
	return nil, gopi.ErrNotImplemented
}

func (this *manager) DestroyBitmap(bitmap gopi.Bitmap) error {
	return gopi.ErrNotImplemented
}
*/

////////////////////////////////////////////////////////////////////////////////
// INTERFACE

func (this *manager) Display() gopi.Display {
	return this.display
}

func (this *manager) Name() string {
	return "TODO"
	//	return fmt.Sprintf("%v %v", rpi.EGLQueryString(this.handle, EGL_VENDOR), rpi.EGLQueryString(this.handle, EGL_VERSION))
}

func (this *manager) Extensions() []string {
	return []string{}
	//	return strings.Split(rpi.EGLQueryString(this.handle, rpi.EGL_EXTENSIONS), " ")
}

// Return capabilities for the GPU
func (this *manager) Types() []gopi.SurfaceType {
	return nil
}

/*
	types := strings.Split(rpi.EGLQueryString(this.handle, rpi.EGL_CLIENT_APIS), " ")
	surface_types := make([]gopi.SurfaceType, 0, 3)
	for _, t := range types {
		if t2, ok := eglStringTypeMap[t]; ok {
			surface_types = append(surface_types, t2)
		}
	}
	// always include RGBA32
	return append(surface_types, gopi.SURFACE_TYPE_RGBA32)

*/
