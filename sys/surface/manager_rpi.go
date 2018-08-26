/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2017
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package surface

import (
	// Frameworks
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type SurfaceManager struct {
	Display gopi.Display
}

type manager struct {
	log          gopi.Logger
	display      gopi.Display
	handle       eglDisplay
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
	n := to_eglNativeDisplayType(this.display.Display())
	if handle, err := eglGetDisplay(n); err != EGL_SUCCESS {
		return nil, os.NewSyscallError("eglGetDisplay", err)
	} else {
		this.handle = handle
	}
	if major, minor, err := eglInitialize(this.handle); err != EGL_SUCCESS {
		return nil, os.NewSyscallError("eglInitialize", err)
	} else {
		this.major = int(major)
		this.minor = int(minor)
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
	if err := eglTerminate(this.handle); err != EGL_SUCCESS {
		return os.NewSyscallError("Close", err)
	}

	// Blank out
	this.display = nil
	this.handle = eglDisplay(EGL_NO_DISPLAY)

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *manager) String() string {
	if this.display == nil {
		return fmt.Sprintf("<graphics.surfacemanager>{ nil }")
	} else {
		return fmt.Sprintf("<graphics.surfacemanager>{ handle=%v name=%v version={ %v,%v } types=%v extensions=%v display=%v }", this.handle, this.Name(), this.major, this.minor, this.Types(), this.Extensions(), this.display)
	}
}

////////////////////////////////////////////////////////////////////////////////
// DO

func (this *manager) Do(callback gopi.SurfaceManagerCallback) error {
	// check parameters
	if this.handle == eglDisplay(EGL_NO_DISPLAY) {
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
	if this.update != dxUpdateHandle(DX_NO_UPDATE) {
		return gopi.ErrOutOfOrder
	}
	if update, err := dxUpdateStart(DX_UPDATE_PRIORITY_DEFAULT); err != DX_SUCCESS {
		return os.NewSyscallError("dxUpdateStart", err)
	} else {
		this.update = update
		return nil
	}
}

func (this *manager) doUpdateEnd() error {
	this.Lock()
	defer this.Unlock()
	if this.update == dxUpdateHandle(DX_NO_UPDATE) {
		return gopi.ErrOutOfOrder
	}
	if err := dxUpdateSubmitSync(this.update); err != DX_SUCCESS {
		return os.NewSyscallError("doUpdateEnd", err)
	} else {
		this.update = dxUpdateHandle(DX_NO_UPDATE)
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

////////////////////////////////////////////////////////////////////////////////
// INTERFACE

func (this *manager) Display() gopi.Display {
	return this.display
}

func (this *manager) Name() string {
	return fmt.Sprintf("%v %v", eglQueryString(this.handle, EGL_VENDOR), eglQueryString(this.handle, EGL_VERSION))
}

func (this *manager) Extensions() []string {
	return strings.Split(eglQueryString(this.handle, EGL_EXTENSIONS), " ")
}

// Return capabilities for the GPU
func (this *manager) Types() []gopi.SurfaceType {
	types := strings.Split(eglQueryString(this.handle, EGL_CLIENT_APIS), " ")
	surface_types := make([]gopi.SurfaceType, 0, 3)
	for _, t := range types {
		if t2, ok := eglStringTypeMap[t]; ok {
			surface_types = append(surface_types, t2)
		}
	}
	// always include RGBA32
	return append(surface_types, gopi.SURFACE_TYPE_RGBA32)
}
