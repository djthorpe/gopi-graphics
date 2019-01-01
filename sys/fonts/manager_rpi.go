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
	"os"
	"path/filepath"
	"sync"

	// Frameworks
	"github.com/djthorpe/gopi"
	ft "github.com/djthorpe/gopi-hw/freetype"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type FontManager struct{}

type manager struct {
	log                 gopi.Logger
	library             ft.FT_Library
	major, minor, patch int
	faces               map[string]gopi.FontFace
	sync.Mutex
}

type face struct {
	handle ft.FT_Face
	path   string
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config FontManager) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<graphics.fonts.Open>{ }")

	this := new(manager)
	this.log = log
	this.faces = make(map[string]gopi.FontFace, 0)

	this.Lock()
	defer this.Unlock()

	if library, err := ft.FT_Init(); err != nil {
		return nil, err
	} else {
		this.library = library
		this.major, this.minor, this.patch = ft.FT_Library_Version(this.library)
	}

	return this, nil
}

func (this *manager) Close() error {
	this.log.Debug("<graphics.fonts.Close>{ handle=0x%X }", this.library)

	this.Lock()
	defer this.Unlock()

	if this.library == nil {
		return nil
	}

	for k, face := range this.faces {
		if err := this.DestroyFace(face); err != nil {
			return err
		} else {
			delete(this.faces, k)
		}
	}

	if err := ft.FT_Destroy(this.library); err != nil {
		return err
	}

	// Release resources
	this.library = nil
	this.faces = nil
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *manager) String() string {
	return fmt.Sprintf("<graphics.fonts.Manager>{ handle=0x%X version={%v,%v,%v} }", this.library, this.major, this.minor, this.patch)
}

////////////////////////////////////////////////////////////////////////////////
// INTERFACE

func (this *manager) OpenFace(path string) (gopi.FontFace, error) {
	return this.OpenFaceAtIndex(path, 0)
}

func (this *manager) OpenFaceAtIndex(path string, index uint) (gopi.FontFace, error) {
	this.log.Debug2("<graphics.fonts.OpenFaceAtIndex{ path=%v index=%v }", path, index)

	// Create the face
	face := &face{
		path: filepath.Clean(path),
	}

	this.Lock()
	defer this.Unlock()

	if handle, err := ft.FT_NewFace(this.library, path, index); err != nil {
		return nil, err
	} else if err := ft.FT_SelectCharmap(handle, ft.FT_ENCODING_UNICODE); err != nil {
		ft.FT_DoneFace(handle)
		return nil, err
	} else {
		face.handle = handle
	}

	// VG Create Font
	//face.font = C.vgCreateFont(C.VGint(face.GetNumGlyphs()))
	//if face.font == VG_FONT_NONE {
	//	this.vgfontDoneFace(face.handle)
	//	return nil, vgGetError(vgErrorType(C.vgGetError()))
	//}

	// Load Glyphs
	//if err := this.LoadGlyphs(face, 64.0, 0.0); err != nil {
	//	this.vgfontDoneFace(face.handle)
	//	C.vgDestroyFont(face.font)
	//	return nil, err
	//}

	// Add face to list of faces
	this.faces[face.path] = face

	return face, nil
}

func (this *manager) OpenFacesAtPath(path string, callback func(manager gopi.FontManager, path string, info os.FileInfo) bool) error {
	this.log.Debug2("<graphics.fonts.OpenFacesAtPath{ path=%v }", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if callback(this, path, info) == false {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}
		if info.IsDir() {
			return nil
		}
		// Open zero-indexed face
		face, err := this.OpenFace(path)
		if err != nil {
			return err
		}
		// If there are more faces in the file, then load these too
		if face.NumFaces() > uint(1) {
			for i := uint(1); i < face.NumFaces(); i++ {
				_, err := this.OpenFaceAtIndex(path, i)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	return err
}

func (this *manager) DestroyFace(f gopi.FontFace) error {
	this.log.Debug2("<graphics.fonts.DestroyFace{ face=%v }", f)
	if face_, ok := f.(*face); ok == false {
		return gopi.ErrBadParameter
	} else if _, exists := this.faces[face_.path]; exists == false {
		return gopi.ErrBadParameter
	} else {
		delete(this.faces, face_.path)
		return ft.FT_DoneFace(face_.handle)
	}
}

func (this *manager) FaceForPath(path string) gopi.FontFace {
	if face, exists := this.faces[filepath.Clean(path)]; exists {
		return face
	} else {
		return nil
	}
}

func (this *manager) Families() []string {
	families := make(map[string]bool, 0)
	for _, face := range this.faces {
		family := face.Family()
		if _, exists := families[family]; exists {
			continue
		}
		families[family] = true
	}
	familes_ := make([]string, 0, len(families))
	for k := range families {
		familes_ = append(familes_, k)
	}
	return familes_
}

func (this *manager) Faces(family string, flags gopi.FontFlags) []gopi.FontFace {
	faces := make([]gopi.FontFace, 0)
	for _, face := range this.faces {
		if family != "" && family != face.Family() {
			continue
		}
		switch flags {
		case gopi.FONT_FLAGS_STYLE_ANY:
			faces = append(faces, face)
		case gopi.FONT_FLAGS_STYLE_REGULAR, gopi.FONT_FLAGS_STYLE_BOLD, gopi.FONT_FLAGS_STYLE_ITALIC, gopi.FONT_FLAGS_STYLE_BOLDITALIC:
			if face.Flags()&flags == flags {
				faces = append(faces, face)
			}
		}
	}
	return faces
}
