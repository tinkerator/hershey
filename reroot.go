//go:build go1.24
// +build go1.24

package hershey

import "os"

// NewFontDir replaces the set of font directory content with a new
// set from a path. Unopened fonts from the prior font directories are
// discarded.
func NewFontDir(path string) error {
	if path == "" {
		workingFS = embeddedJHF
		fontDir = defaultFontDirSearch
		barePath = defaultFontDir
		return nil
	}
	root, err := os.OpenRoot(path)
	if err != nil {
		return err
	}
	fontDir = ""
	barePath = "."
	workingFS = root.FS()
	return nil
}
