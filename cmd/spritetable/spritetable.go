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
			z := (y%3 + x) % 3
			drawBG(img, s, t, 16, 16, z)
			d.Dot = fixed.Point26_6{
				X: fixed.Int26_6(s << 6),
				Y: fixed.Int26_6((t-1)<<6) + m.Ascent,
			}
			d.DrawString(fmt.Sprintf("%X%x", y, x))
		}
	}

	f, err := os.Create("spritetable.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

var bgcolors = []color.RGBA{
	{128, 0, 0, 255},
	{0, 128, 0, 255},
	{0, 0, 128, 255},
}

func drawBG(img *image.RGBA, x, y, w, h, z int) {
	c := bgcolors[z]
	for i := 2; i < w-2; i++ {
		for j := 2; j < h-2; j++ {
			img.Set(x+i, y+j, c)
		}
	}
}
