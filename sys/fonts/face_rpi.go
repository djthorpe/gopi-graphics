/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2018
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package fonts

import (
	"fmt"
	"path"

	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
  #cgo CFLAGS:   -I/usr/include/freetype2
  #cgo LDFLAGS:  -lfreetype
  #include <ft2build.h>
  #include FT_FREETYPE_H
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS: Face information

func (this *face) String() string {
	return fmt.Sprintf("<graphics.fonts.Face>{ name=%v index=%v family=%v style=%v num_faces=%v num_glyphs=%v }", this.Name(), this.Index(), this.Family(), this.Style(), this.NumFaces(), this.NumGlyphs())
}

func (this *face) Name() string {
	return path.Base(this.path)
}

func (this *face) Family() string {
	return C.GoString((*C.char)(this.handle.family_name))
}

func (this *face) Style() string {
	return C.GoString((*C.char)(this.handle.style_name))
}

func (this *face) Index() uint {
	return uint(this.handle.face_index)
}

func (this *face) NumFaces() uint {
	return uint(this.handle.num_faces)
}

func (this *face) NumGlyphs() uint {
	return uint(this.handle.num_glyphs)
}

func (this *face) Flags() gopi.FontFlags {
	return gopi.FontFlags(this.handle.style_flags)
}
