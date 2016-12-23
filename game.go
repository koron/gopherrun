package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"golang.org/x/image/math/fixed"
)

var (
	// scw is screen character width
	scw = (screenWidth+cellWidth-1)/cellWidth + 1
	// sch is screen character height
	sch    = (screenHeight+cellHeight-1)/cellHeight + 1
	// maxBgOffx is max value for bgOffX
	maxBgOffx = fixed.I(16)
)

type Game struct {
	win *sdl.Window
	ren *sdl.Renderer
	ch1 *sdl.Texture
	ch2 *sdl.Texture

	running  bool
	frameNum uint64
	pressedA bool
	pressedB bool

	bgMap  []uint8
	bgOffX fixed.Int26_6
	bgOffY fixed.Int26_6

	speedX fixed.Int26_6
}

// Init initialize all game status.
func (g *Game) Init() error {
	g.bgMap = make([]uint8, scw*sch)
	for x := 0; x < scw; x++ {
		for y := 10; y < sch; y++ {
			g.bgMap[x*sch+y] = uint8(x%16 + 16)
		}
	}
	g.running = true
	g.speedX = fixed.I(1) / 3
	return nil
}

func (g *Game) Run() error {
	for g.running {
		g.initNewFrame()
		for ev := sdl.PollEvent(); ev != nil; ev = sdl.PollEvent() {
			g.procEvent(ev)
		}
		g.update()
		if err := g.render(); err != nil {
			return err
		}
		g.ren.Present()
	}
	return nil
}

func (g *Game) drawBG() error {
	i := 0
	src := sdl.Rect{0, 0, int32(cellWidth), int32(cellHeight)}
	dst := sdl.Rect{0, 0, int32(cellWidth), int32(cellHeight)}
	for x := 0; x < scw; x++ {
		dst.X = int32(x*cellWidth - g.bgOffX.Floor())
		for y := 0; y < sch; y++ {
			n := int(g.bgMap[i])
			dst.Y = int32(y*cellHeight - g.bgOffY.Floor())
			src.X = int32((n % 16) * cellWidth)
			src.Y = int32((n / 16) * cellWidth)
			if err := g.ren.Copy(g.ch1, &src, &dst); err != nil {
				return err
			}
			i++
		}
	}
	return nil
}

// initNewFrame initialize state for a new frame.
func (g *Game) initNewFrame() {
	g.frameNum++
	g.pressedA = false
	g.pressedB = false
}

func (g *Game) procEvent(raw sdl.Event) {
	switch ev := raw.(type) {
	case *sdl.KeyDownEvent:
		if ev.Repeat != 0 {
			break
		}
		switch ev.Keysym.Sym {
		case sdl.K_SPACE, sdl.K_RETURN:
			g.pressedA = true
		case sdl.K_LSHIFT, sdl.K_RSHIFT, sdl.K_LCTRL, sdl.K_RCTRL:
			g.pressedB = true
		case sdl.K_ESCAPE:
			g.running = false
		}
	}
}

func (g *Game) update() {
	g.bgOffX += g.speedX
	for g.bgOffX >= maxBgOffx {
		g.bgOffX -= maxBgOffx
		g.shiftBG()
	}
}

func (g *Game) shiftBG() {
	l := len(g.bgMap)
	copy(g.bgMap[0:l-sch], g.bgMap[sch:])
}

func (g *Game) render() error {
	if err := g.drawBG(); err != nil {
		return err
	}
	return nil
}
