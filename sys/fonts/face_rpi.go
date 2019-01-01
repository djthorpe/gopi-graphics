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
	ft "github.com/djthorpe/gopi-hw/freetype"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS: Face information

func (this *face) String() string {
	return fmt.Sprintf("<graphics.fonts.Face>{ name=%v index=%v family=%v style=%v num_faces=%v num_glyphs=%v }", this.Name(), this.Index(), this.Family(), this.Style(), this.NumFaces(), this.NumGlyphs())
}

func (this *face) Name() string {
	return path.Base(this.path)
}

func (this *face) Family() string {
	return ft.FT_FaceFamily(this.handle)
}

func (this *face) Style() string {
	return ft.FT_FaceStyle(this.handle)
}

func (this *face) Index() uint {
	return ft.FT_FaceIndex(this.handle)
}

func (this *face) NumFaces() uint {
	return ft.FT_FaceNumFaces(this.handle)
}

func (this *face) NumGlyphs() uint {
	return ft.FT_FaceNumGlyphs(this.handle)
}

func (this *face) Flags() gopi.FontFlags {
	return ft.FT_FaceStyleFlags(this.handle)
}
