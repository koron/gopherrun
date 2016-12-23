package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

// generate character table
func main() {
	img := image.NewRGBA(image.Rect(0, 0, 256, 256))

	face := inconsolata.Regular8x16
	m := face.Metrics()
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{255, 255, 255, 255}),
		Face: face,
	}
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			s, t := x*16, y*16
			drawBG(img, s, t, 16, 16)
			d.Dot = fixed.Point26_6{
				fixed.Int26_6(s << 6),
				fixed.Int26_6((t-1)<<6) + m.Ascent,
			}
			d.DrawString(fmt.Sprintf("%X%x", y, x))
		}
	}

	f, err := os.Create("chartable.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

func drawBG(img *image.RGBA, x, y, w, h int) {
	fg := color.RGBA{128, 0, 0, 255}
	bg := color.RGBA{0, 0, 0, 255}
	for i := 0; i < w; i++ {
		img.Set(x+i, y, fg)
		img.Set(x+i, y+h-1, fg)
	}
	for i := 1; i < h-1; i++ {
		img.Set(x, y+i, fg)
		img.Set(x+w-1, y+i, fg)
		for j := 1; j < w-1; j++ {
			img.Set(x+j,y+i, bg)
		}
	}
}
