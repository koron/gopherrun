package main

import (
	"log"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

var renderFlags uint32 = sdl.RENDERER_ACCELERATED | sdl.RENDERER_PRESENTVSYNC

var (
	screenWidth  = 320
	screenHeight = 180
	cellWidth    = 16
	cellHeight   = 16
)

func loadTexture(r *sdl.Renderer, name string) (*sdl.Texture, *sdl.Surface, error) {
	s, err := img.Load(name)
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
	t1, s1, err := loadTexture(r, "chartable.png")
	if err != nil {
		return err
	}
	defer t1.Destroy()
	defer s1.Free()

	// sprite characters
	t2, s2, err := loadTexture(r, "spritetable.png")
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

	m1, err := mix.LoadMUS("jump07.mp3")
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
	if err := runGame(); err != nil {
		log.Fatal(err)
	}
}
