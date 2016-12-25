package main

import (
	"math/rand"

	"github.com/veandco/go-sdl2/sdl"
	mix "github.com/veandco/go-sdl2/sdl_mixer"
	"golang.org/x/image/math/fixed"
)

var (
	// scw is screen character width
	scw = (screenWidth+cellWidth-1)/cellWidth + 1

	// sch is screen character height
	sch = (screenHeight+cellHeight-1)/cellHeight + 1

	// maxBgOffx is max value for bgOffX
	maxBgOffx = fixed.I(16)

	// gravityPower
	gravityPower = fixed.I(3) / 2

	maxSpeedY = fixed.I(16) / 2

	// risingPower
	risingPower = fixed.I(11) / 2

	risingInitN = 10

	gopherX = fixed.I(320) / 5

	initSpeedX = fixed.I(1) / 3

	maxSpeedX = fixed.I(6) / 2

	accelX = fixed.I(1) / 40

	gopherInitY = fixed.I(8 * 16)

	walkPattern = []int{3, 4, 5}
)

type Game struct {
	win *sdl.Window
	ren *sdl.Renderer
	ch1 *sdl.Texture
	ch2 *sdl.Texture
	se1 *mix.Music

	running   bool
	frameNum  uint64
	pressedA  bool
	releasedA bool
	pressedB  bool

	bgMap  []uint8
	bgOffX fixed.Int26_6
	bgOffY fixed.Int26_6

	spPatterns []SpritePattern
	sprites    []Sprite

	mode     Mode
	gopherY  fixed.Int26_6
	speedX   fixed.Int26_6
	speedY   fixed.Int26_6
	floating bool
	risingN  int

	animeIndex int
	animeFrame int

	rand         *rand.Rand
	groundHeight int
	groundHole   bool
	groundCont   int
}

type Sprite struct {
	id int
	x  int32
	y  int32
}

type SpritePattern struct {
	x int32
	y int32
	w int32
	h int32
}

type Mode int

const (
	title Mode = iota
	playing
	gameover
)

// Init initialize all game status.
func (g *Game) Init() error {
	g.bgMap = make([]uint8, scw*sch)
	g.spPatterns = []SpritePattern{
		SpritePattern{x: 0, y: 0, w: 16, h: 32},
		SpritePattern{x: 16, y: 0, w: 16, h: 32},
		SpritePattern{x: 32, y: 0, w: 16, h: 32},
		SpritePattern{x: 48, y: 0, w: 16, h: 32},
		SpritePattern{x: 64, y: 0, w: 16, h: 32},
		SpritePattern{x: 80, y: 0, w: 16, h: 32},
		SpritePattern{x: 96, y: 0, w: 16, h: 32},
	}
	g.sprites = []Sprite{
		Sprite{id: 0, x: int32(gopherX.Floor()), y: 0},
	}
	g.gotoTitle()
	return nil
}

func (g *Game) gotoTitle() {
	for x := 0; x < scw; x++ {
		for y := 0; y < 10; y++ {
			g.bgMap[x*sch+y] = 0x00
		}
		for y := 10; y < sch; y++ {
			g.bgMap[x*sch+y] = 0x10
		}
	}
	g.bgOffX = 0
	g.mode = title
	g.gopherY = gopherInitY
	g.running = true
	g.speedX = initSpeedX
	g.speedY = 0
	g.floating = false
	g.risingN = 0
	g.animeIndex = 0
	g.animeFrame = 0
	g.rand = rand.New(rand.NewSource(114514))
	g.groundHeight = 10
	g.groundHole = false
	g.groundCont = 5
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
	g.releasedA = false
	g.pressedB = false
}

func (g *Game) procEvent(raw sdl.Event) {
	switch ev := raw.(type) {
	case *sdl.WindowEvent:
		if ev.Event == sdl.WINDOWEVENT_CLOSE {
			g.running = false
		}
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
	case *sdl.KeyUpEvent:
		switch ev.Keysym.Sym {
		case sdl.K_SPACE, sdl.K_RETURN:
			g.releasedA = true
		}
	}
}

func (g *Game) shiftBG() {
	l := len(g.bgMap)
	copy(g.bgMap[0:l-sch], g.bgMap[sch:])
}

func (g *Game) drawSprites() error {
	for i := len(g.sprites) - 1; i >= 0; i-- {
		s := g.sprites[i]
		p := g.spPatterns[s.id]
		src := sdl.Rect{p.x, p.y, p.w, p.h}
		dst := sdl.Rect{s.x, s.y, p.w, p.h}
		if err := g.ren.Copy(g.ch2, &src, &dst); err != nil {
			return err
		}
	}
	return nil
}

func (g *Game) render() error {
	if err := g.drawBG(); err != nil {
		return err
	}
	if err := g.drawSprites(); err != nil {
		return err
	}
	return nil
}

func (g *Game) update() {
	if g.floating {
		if g.risingN > 0 {
			if g.releasedA {
				g.risingN = 0
			} else {
				g.risingN--
			}
			g.speedY = -risingPower
		} else {
			g.speedY += gravityPower
			if g.speedY > maxSpeedY {
				g.speedY = maxSpeedY
			}
		}
	} else {
		if g.pressedA {
			g.floating = true
			g.risingN = risingInitN
			g.speedY = -risingPower
			g.mode = playing
			g.se1.Play(1)
		}
	}

	switch g.mode {
	case title:
		g.speedX = initSpeedX
	case playing:
		g.speedX += accelX
		if g.speedX > maxSpeedX {
			g.speedX = maxSpeedX
		}
	case gameover:
		g.speedX = 0
	}

	g.bgOffX += g.speedX
	// check to hit wall
	if g.speedX > 0 {
		y := g.gopherY.Floor()
		cx := ((gopherX + g.bgOffX).Floor() + cellWidth) / cellWidth
		cy := y / cellWidth
		ch := 3
		if y%cellHeight == 0 {
			ch = 2
		}
		hit := false
		for i := 0; i < ch; i++ {
			cy2 := cy + i
			if cy2 < 0 {
				continue
			} else if cy2 >= sch {
				break
			}
			if g.bgMap[cx*sch+cy+i] >= 0x10 {
				hit = true
				break
			}
		}
		if hit {
			g.speedX = 0
			g.bgOffX = fixed.I((cx-1)*cellWidth) - gopherX
		}
	}

	// scroll and prepare new area
	for g.bgOffX >= maxBgOffx {
		g.bgOffX -= maxBgOffx
		g.shiftBG()
		// insert new bgMap at right
		switch g.mode {
		case title:
			n := (scw - 1) * sch
			for y := 0; y < sch; y++ {
				if y < 10 {
					g.bgMap[n+y] = 0x00
				} else {
					g.bgMap[n+y] = 0x10
				}
			}
		case playing:
			// TODO: generate stage data
			n := (scw - 1) * sch
			for y := 0; y < sch; y++ {
				if !g.groundHole && y >= g.groundHeight {
					g.bgMap[n+y] = 0x10
				} else {
					g.bgMap[n+y] = 0x00
				}
			}
			g.groundCont--
			if g.groundCont <= 0 {
				if !g.groundHole && g.rand.Float32() < 0.17 {
					c := int(g.rand.ExpFloat64() * 1.5)
					if c < 1 {
						c = 1
					} else if c > 4 {
						c = 4
					}
					g.groundHole = true
					g.groundCont = c
				} else {
					if r := g.rand.Float32(); r < 0.18 {
						c := int(g.rand.ExpFloat64() * 1)
						if c < 1 {
							c = 1
						} else if c > 4 {
							c = 4
						}
						g.groundHeight -= c
						if g.groundHeight < 4 {
							g.groundHeight = 4
						}
					} else if r >= 0.82 {
						c := int(g.rand.ExpFloat64() * 1)
						if c < 1 {
							c = 1
						} else if c > 4 {
							c = 4
						}
						g.groundHeight += c
						if g.groundHeight > 10 {
							g.groundHeight = 10
						}
					}
					c := int(g.rand.NormFloat64()*2 + 3)
					if c < 1 {
						c = 1
					} else if c > 8 {
						c = 8
					}
					g.groundHole = false
					g.groundCont = c
				}
			}
		}
	}

	g.gopherY += g.speedY
	// check to touch grand
	if g.speedY >= 0 {
		x := (gopherX + g.bgOffX).Floor()
		cx := x / cellWidth
		cy := (g.gopherY.Floor() + cellHeight*2) / cellHeight
		cw := 2
		if x%cellWidth == 0 {
			cw = 1
		}
		if cy >= 0 && cy < sch {
			touch := false
			for i := 0; i < cw; i++ {
				if g.bgMap[(cx+i)*sch+cy] >= 0x10 {
					touch = true
					break
				}
			}
			if touch {
				g.gopherY = fixed.I((cy - 2) * cellHeight)
				g.speedY = 0
				g.floating = false
				g.risingN = 0
			} else {
				g.floating = true
			}
		}
	}
	if g.gopherY.Floor() > screenHeight {
		// game over
		g.mode = gameover
		// FIXME: show game over message
	}

	if g.mode == gameover && g.pressedA {
		// back to title
		g.gotoTitle()
	}

	if g.floating {
		g.animeIndex = 6
		g.animeFrame = 0
	} else if g.speedX > 0 {
		g.animeIndex = walkPattern[g.animeFrame/10]
		g.animeFrame++
		if g.animeFrame >= len(walkPattern)*10 {
			g.animeFrame = 0
		}
	} else {
		g.animeFrame = 0
		g.animeIndex = 0
	}

	g.sprites[0].y = int32(g.gopherY.Floor())
	g.sprites[0].id = g.animeIndex
}
