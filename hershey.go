// Package hershey provides an API to obtain vector data to render
// Hershey fonts. The defaults for this package load the curated
// version of the fonts stored in jhf format.
package hershey

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"sync"
)

//go:embed fonts/*.jhf
var embeddedJHF embed.FS

// workingFS points to the fs.FS interface that we read files from.
var workingFS fs.FS = embeddedJHF

const defaultFontDir = "fonts"
const defaultFontDirSearch = defaultFontDir + "/"

var fontDir = defaultFontDirSearch
var barePath = defaultFontDir

// Font holds loaded font data. It caches the decoding output of used
// Glyphs.
type Font struct {
	data    map[int]string
	mu      sync.Mutex
	decoded map[int]Glyph
}

// Glyph holds the coordinates of a Glyph, the Left etc values indicate
// the bounding box of the Glyph.
type Glyph struct {
	// Left etc hold the bottom left and top right coordinates of
	// the Glyph. In the form that the page lines are more
	// positive the further down the page you go. Top and Bottom
	// are derived from the Strokes, Left and Right are from the
	// Glyph's encoding.
	Left, Bottom, Right, Top int
	// Strokes hold the row,col coordinates of the points in
	// penned lines. There may be more than one line in a Glyph.
	Strokes [][][2]int
}

// List lists the names of the known fonts.
func List() []string {
	fs, err := fs.ReadDir(workingFS, barePath)
	if err != nil {
		return nil
	}
	var ls []string
	for _, f := range fs {
		if name := f.Name(); strings.HasSuffix(name, ".jhf") {
			ls = append(ls, strings.TrimSuffix(strings.TrimPrefix(name, fontDir), ".jhf"))
		}
	}
	return ls
}

// New unpacks a named font if known.
func New(name string) (*Font, error) {
	path := fmt.Sprint(fontDir, name, ".jhf")
	fs, err := fs.ReadFile(workingFS, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q (%q)", name, path)
	}
	font := make(map[int]string)
	autoInc := 1
	pending := ""
	for i, f := range bytes.Split(fs, []byte("\n")) {
		pending = pending + string(f)
		const col = 8
		if pending == "" {
			continue // typically the last line
		}
		if len(pending) < col {
			return nil, fmt.Errorf("illegal size for entry %d (%q)", i, pending)
		}
		toks := strings.Fields(pending[5:8])
		if len(toks) != 1 {
			return nil, fmt.Errorf("parse issue for %q entry %d, %q", name, i, pending[5:8])
		}
		chs, err := strconv.Atoi(toks[0])
		if err != nil {
			return nil, fmt.Errorf("%q entry %d: %v", name, i, err)
		}
		if want, got := 2*chs, len(pending[col:]); want < got {
			return nil, fmt.Errorf("%q entry %d: excessive data (%d > %d) %q", name, i, got, want, pending[col:])
		} else if got < want {
			pending = pending
			continue
		}
		toks = strings.Fields(string(pending[:5]))
		if len(toks) != 1 {
			return nil, fmt.Errorf("parse issue for %q entry %d, %q", name, i, pending[:5])
		}
		n, err := strconv.Atoi(toks[0])
		if err != nil {
			return nil, fmt.Errorf("%q entry %d has bad start value, %q: %v", name, i, toks[0], err)
		}
		if n == 12345 {
			n += autoInc
			autoInc++
		}
		font[n] = pending[col:]
		pending = ""
	}
	return &Font{
		data:    font,
		decoded: make(map[int]Glyph),
	}, nil
}

// Strokes returns the Glyph associated with index from this font.
func (font *Font) Strokes(index int) (Glyph, error) {
	font.mu.Lock()
	defer font.mu.Unlock()
	gl, ok := font.decoded[index]
	if ok {
		return gl, nil
	}
	d, ok := font.data[index]
	if !ok {
		return gl, fmt.Errorf("glyph code %d unknown", index)
	}
	left, right := int(d[0])-int('R'), int(d[1])-int('R')
	var strokes [][][2]int
	var scribe [][2]int
	var minY, maxY int
	for i := 2; i <= len(d); i += 2 {
		if i == len(d) || d[i:i+2] == " R" {
			strokes = append(strokes, scribe)
			scribe = nil
			continue
		}
		x, y := int(d[i])-int('R'), int(d[i+1])-int('R')
		scribe = append(scribe, [2]int{x, y})
		if i == 2 || y <= minY {
			minY = y - 1
		}
		if i == 2 || y >= maxY {
			maxY = y + 1
		}
	}
	gl.Left = left
	gl.Right = right
	gl.Bottom = maxY
	gl.Top = minY
	gl.Strokes = strokes

	font.decoded[index] = gl
	return gl, nil
}

// Scan returns the numerically sorted indices of a font in the form
// of a channel that is closed after all have been read.
func (font *Font) Scan() <-chan int {
	indices := make(chan int)
	go func() {
		defer close(indices)
		var is []int
		for i := range font.data {
			is = append(is, i)
		}
		sort.Ints(is)
		for _, i := range is {
			indices <- i
		}
	}()
	return indices
}

// Marshal encodes a Glyph into its stored format.
func (gl Glyph) Marshal(code int) string {
	enc := func(x, y int) string {
		cx := rune(int('R') + x)
		cy := rune(int('R') + y)
		return fmt.Sprintf("%c%c", cx, cy)
	}
	pairs := []string{enc(gl.Left, gl.Right)}
	for i, lines := range gl.Strokes {
		if i != 0 {
			pairs = append(pairs, " R")
		}
		for _, pt := range lines {
			pairs = append(pairs, enc(pt[0], pt[1]))
		}
	}
	var parts []string
	text := fmt.Sprintf("%5d%3d%s", code, len(pairs), strings.Join(pairs, ""))
	for i := 0; i < len(text); i += 72 {
		if d := len(text) - i; d < 72 {
			if d > 0 {
				parts = append(parts, text[i:])
			}
			break
		}
		parts = append(parts, text[i:i+72])
	}
	return strings.Join(parts, "\n")
}
