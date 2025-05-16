// Program query provides a simple command line tool to demonstrate
// features of the hershey package.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"zappem.net/pub/graphics/hershey"
)

var (
	font    = flag.String("font", "rowmans", "name of font")
	ls      = flag.Bool("ls", false, "list known fonts and exit")
	scan    = flag.Bool("scan", false, "list known codes for --font")
	glyph   = flag.Int("glyph", -1, "display font code as ascii art")
	dir     = flag.String("dir", "", "font directory, Ex. jhfdata/hershey")
	xlate   = flag.String("xlate", "", "utf8 translation files")
	skipped = flag.Bool("skipped", false, "just show codes without utf8 translation")
	dest    = flag.String("dest", "", "write translated *.jhf files here")
	banner  = flag.String("banner", "", "text to display")
)

// scribe renders a width 1 pixel gray line from (x0,y0) to (x1,y1) on im.
func scribe(im *image.Gray16, x0, y0, x1, y1 int) {
	if x1 < x0 {
		x0, y0, x1, y1 = x1, y1, x0, y0
	}
	dx, dy := x1-x0, y1-y0
	inc := 1
	if dy < 0 {
		dy = -dy
		inc = -1
	}
	im.Set(x1, y1, color.Gray{1})
	if dx > dy {
		for x := x0; x != x1; x++ {
			y := y0 + inc*(x-x0)*dy/dx
			im.Set(x, y, color.Gray{1})
		}
		return
	}
	for y := y0; y != y1; y += inc {
		x := x0 + inc*(y-y0)*dx/dy
		im.Set(x, y, color.Gray{1})
	}
	return
}

// loadXlate loads the translation files (*.utf8) for fonts.
func loadXlate(path string) map[int]int {
	d, err := os.ReadFile(path)
	if err != nil {
		return nil // no translation file
	}
	xlt := make(map[int]int)
	for i, line := range bytes.Split(d, []byte("\n")) {
		var text string
		if first := bytes.Index(line, []byte("#")); first != -1 {
			text = string(line[:first])
		} else {
			text = string(line)
		}
		fields := strings.Fields(text)
		switch len(fields) {
		case 0:
			continue
		case 2:
		default:
			log.Fatalf("file %q contains syntax error on line %d: %q", path, i, fields)
		}
		target, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Fatalf("file %q contains bad target number %q on line %d: %v", path, fields[1], i, err)
		}
		rnge := strings.Split(fields[0], "-")
		from, err := strconv.Atoi(rnge[0])
		if err != nil {
			log.Fatalf("file %q has bad base %q in %q: %v", path, rnge[0], fields[0], err)
		}
		if len(rnge) == 2 {
			to, err := strconv.Atoi(rnge[1])
			if err != nil {
				log.Fatalf("file %q has bad second field %q in %q: %v", path, rnge[1], fields[0], err)
			}
			for n := from; n <= to; n++ {
				xlt[n] = target
				target++
			}
			continue
		}
		xlt[from] = target
	}
	return xlt
}

func render(im *image.Gray16) {
	bbox := im.Bounds()
	base, bottom := bbox.Min.Y, bbox.Max.Y
	left, right := bbox.Min.X, bbox.Max.X
	for y := base; y <= bottom; y++ {
		if y == base {
			fmt.Print(" ")
			for x := left; x <= right; x++ {
				if x == 0 {
					fmt.Print("V")
				} else {
					fmt.Print(" ")
				}
			}
			fmt.Println()
		}
		if y == 0 {
			fmt.Print(">")
		} else {
			fmt.Print(" ")
		}
		for x := left; x <= right; x++ {
			pt := im.Gray16At(x, y)
			if pt.Y != 0 {
				fmt.Print("#")
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println()
	}
}

func show(ft *hershey.Font, xl map[int]int, gl int) {
	detail, err := ft.Strokes(gl)
	if err != nil {
		log.Fatalf("font %q glyph %d problem: %v", *font, gl, err)
	}
	if xl == nil {
		fmt.Printf("\nglyph %d: (%d,%d), (%d,%d)\n", gl, detail.Left, detail.Top, detail.Right, detail.Bottom)
	} else if x, ok := xl[gl]; ok {
		if *skipped {
			return
		}
		fmt.Printf("\nglyph %d [%d = `%s`]: (%d,%d), (%d,%d)\n", gl, x, string(rune(x)), detail.Left, detail.Top, detail.Right, detail.Bottom)
	} else {
		fmt.Printf("\nglyph %d [???]: (%d,%d), (%d,%d)\n", gl, detail.Left, detail.Top, detail.Right, detail.Bottom)
	}
	base, left, bottom := 0, 0, 0
	if base > detail.Top {
		base = detail.Top
	}
	if left > detail.Left {
		left = detail.Left
	}
	if bottom < detail.Bottom {
		bottom = detail.Bottom
	}
	im := image.NewGray16(image.Rect(left, base, detail.Right+1, bottom))
	for _, line := range detail.Strokes {
		var old [2]int
		for i, cs := range line {
			if i == 0 {
				scribe(im, cs[0], cs[1], cs[0], cs[1])
			} else {
				scribe(im, old[0], old[1], cs[0], cs[1])
			}
			old = cs
		}
	}
	render(im)
}

func main() {
	flag.Parse()

	if *dir != "" {
		if err := hershey.NewFontDir(*dir); err != nil {
			log.Fatalf("invalid font directory %q: %v", *dir, err)
		}
	}
	if *ls {
		names := hershey.List()
		log.Printf("known fonts: %q", names)
		return
	}
	var xl map[int]int
	if *xlate != "" {
		xl = loadXlate(fmt.Sprint(*xlate, "/", *font, ".utf8"))
	}
	ft, err := hershey.New(*font)
	if err != nil {
		log.Fatalf("bad font %q: %v", *font, err)
	}

	if *scan {
		for code := range ft.Scan() {
			if *glyph == -1 {
				show(ft, xl, code)
			} else if xl == nil {
				log.Printf("glyph: %d", code)
			} else if x, ok := xl[code]; ok {
				if *skipped {
					continue
				}
				log.Printf("glyph: %d [%d = `%s`]", code, x, string(rune(code)))
			} else {
				log.Printf("glyph: %d [???]", code)
			}
		}
	} else if *glyph != -1 {
		show(ft, xl, *glyph)
		return
	}

	if *banner != "" {
		lookup := make(map[int]int)
		if xl != nil {
			for k, v := range xl {
				lookup[v] = k
			}
		}
		composite, xL, xR := ft.Text(*banner)
		im := image.NewGray16(image.Rect(xL-1, composite.Top-1, xR+1, composite.Bottom+1))
		for _, line := range composite.Strokes {
			var old [2]int
			for i, cs := range line {
				if i == 0 {
					scribe(im, cs[0], cs[1], cs[0], cs[1])
				} else {
					scribe(im, old[0], old[1], cs[0], cs[1])
				}
				old = cs
			}
		}
		render(im)
	}

	if *dest != "" {
		if xl == nil {
			log.Fatalf("no translation --xlate to perform into --dest=%q directoty", *dest)
		}

		gls := make(map[int]hershey.Glyph)
		for k, v := range xl {
			gl, err := ft.Strokes(k)
			if err != nil {
				log.Fatalf("unable to find translation for %d (%d) in %q: %v", k, v, *font, err)
			}
			gls[v] = gl
		}
		var us []int
		for k := range gls {
			us = append(us, k)
		}
		sort.Ints(us)
		path := fmt.Sprintf("%s/%s.jhf", *dest, *font)
		f, err := os.Create(path)
		if err != nil {
			log.Fatalf("failed to create %q: %v", path, err)
		}
		defer f.Close()
		for _, c := range us {
			gl := gls[c]
			fmt.Fprintln(f, gl.Marshal(c))
		}
		log.Printf("wrote %q", path)
		return
	}
}
