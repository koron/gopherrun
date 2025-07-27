package main

import (
	_ "embed"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

//go:embed _resources/chartable.png
var chartable []byte

//go:embed _resources/spritetable.png
var spritetable []byte

//go:embed _resources/jump07.mp3
var jumpSound []byte

var renderFlags uint32 = sdl.RENDERER_ACCELERATED | sdl.RENDERER_PRESENTVSYNC

// loadTexture load a texture from memory.
func loadTexture(r *sdl.Renderer, p []byte) (*sdl.Texture, *sdl.Surface, error) {
	rw, err := sdl.RWFromMem(p)
	if err != nil {
		return nil, nil, err
	}
	defer rw.Close()
	s, err := img.LoadRW(rw, false)
	if err != nil {
		return nil, nil, err
	}
	t, err := r.CreateTextureFromSurface(s)
	if err != nil {
		s.Free()
		return nil, nil, err
	}
	return t, s, nil
}

func loadMusic(p []byte) (*mix.Music, error) {
	rw, err := sdl.RWFromMem(p)
	if err != nil {
		return nil, err
	}
	return mix.LoadMUSRW(rw, 1)
}

// runGame setup resources and run a game.
func runGame() error {
	sdl.Init(sdl.INIT_EVERYTHING)
	defer sdl.Quit()

	w, err := sdl.CreateWindow("Gopher Run!", sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED, 1280, 720, sdl.WINDOW_SHOWN)
	if err != nil {
		return err
	}
	defer w.Destroy()

	r, err := sdl.CreateRenderer(w, -1, renderFlags)
	if err != nil {
		return err
	}
	defer r.Destroy()
	r.SetLogicalSize(320, 180)

	// background characters
	t1, s1, err := loadTexture(r, chartable)
	if err != nil {
		return err
	}
	defer t1.Destroy()
	defer s1.Free()

	// sprite characters
	t2, s2, err := loadTexture(r, spritetable)
	if err != nil {
		return err
	}
	defer t2.Destroy()
	defer s2.Free()

	err = mix.OpenAudio(mix.DEFAULT_FREQUENCY, mix.DEFAULT_FORMAT,
		mix.DEFAULT_CHANNELS, mix.DEFAULT_CHUNKSIZE)
	if err != nil {
		return err
	}

	m1, err := loadMusic(jumpSound)
	if err != nil {
		return err
	}
	defer m1.Free()

	// FIXME: setup  more resources

	g := &Game{
		win: w,
		ren: r,
		ch1: t1,
		ch2: t2,
		se1: m1,
	}
	if err := g.Init(); err != nil {
		return err
	}
	return g.Run()
}

func main() {
	//if err := runGame(); err != nil {
	//	log.Fatal(err)
	//}

	var game Game2
	if err := game.Init(); err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Gopher Run!")
	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
