/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016-2018
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

// Outputs a table of displays - works on RPi at the moment
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/olekukonko/tablewriter"

	// Modules
	_ "github.com/djthorpe/gopi-graphics/sys/fonts"
	_ "github.com/djthorpe/gopi/sys/logger"
)

////////////////////////////////////////////////////////////////////////////////

func CheckFont(manager gopi.FontManager, path string, info os.FileInfo) bool {
	if info.IsDir() {
		// Allow subfolders to be walked
		return true
	}
	// Check for face already loaded
	if manager.FaceForPath(path) != nil {
		return false
	}
	// Check file extensions
	if ext := strings.ToLower(filepath.Ext(path)); ext == ".ttf" || ext == ".ttc" || ext == ".otf" || ext == ".otc" {
		return true
	} else {
		return false
	}
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	if app.Fonts == nil {
		return fmt.Errorf("Missing Font Manager")
	}
	if path_flag, exists := app.AppFlags.GetString("font.path"); exists == false {
		return fmt.Errorf("Missing -font.path flag")
	} else {
		// Load all the faces
		for _, path := range strings.Split(path_flag, ":") {
			if stat, err := os.Stat(path); os.IsNotExist(err) || stat.IsDir() == false {
				return fmt.Errorf("Invalid path: %v", path)
			} else if err := app.Fonts.OpenFacesAtPath(path, CheckFont); err != nil {
				return err
			}
		}

		// Output font family information
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Family"})
		for _, family := range app.Fonts.Families() {
			table.Append([]string{family})
		}
		table.Render()

		// Output all fonts
		table2 := tablewriter.NewWriter(os.Stdout)
		table2.SetHeader([]string{"Name", "Index", "Family", "Style", "Flags", "Glyphs"})
		for _, face := range app.Fonts.Faces("", gopi.FONT_FLAGS_STYLE_ANY) {
			table2.Append([]string{
				face.Name(),
				fmt.Sprint(face.Index()),
				face.Family(),
				face.Style(),
				fmt.Sprint(face.Flags()),
				fmt.Sprint(face.NumGlyphs()),
			})
		}
		table2.Render()

	}
	return nil
}

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("fonts")

	// Set the font path
	config.AppFlags.FlagString("font.path", "", "Colon-separated list of font locations")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main))
}
