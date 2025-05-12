//go:build !go1.24
// +build !go1.24

package hershey

import "fmt"

// NewFontDir replaces the set of font directory content with a new
// set from a path. Unopened fonts from the prior font directories are
// discarded.
func NewFontDir(path string) error {
	return fmt.Errorf("upgrade go to 1.24 to set font directory to %q", path)
}
