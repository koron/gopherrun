package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	sdl.Init(sdl.INIT_EVERYTHING)

	window, err := sdl.CreateWindow("Gopher Run!", sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED, 1280, 720, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	w, h := window.GetSize()
	fmt.Printf("Size: %d, %d\n", w, h)

	r, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer r.Destroy()
	r.SetLogicalSize(320, 180)
	r.SetDrawColor(255, 0, 0, 255)
	r.FillRect(&sdl.Rect{0, 0, 200, 200})
	r.Present()

	sdl.Delay(1000)

	r.SetDrawColor(0, 0, 255, 255)
	r.FillRect(&sdl.Rect{200, 100, 200, 200})
	r.Present()

	sdl.Delay(2000)
	sdl.Quit()
}
