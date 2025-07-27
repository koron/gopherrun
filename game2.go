package main

import (
	"embed"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"io/fs"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/math/fixed"
)

//go:embed _resources
var embedFS embed.FS

var resourcesFS fs.FS

func init() {
	fsys, err := fs.Sub(embedFS, "_resources")
	if err != nil {
		panic(err.Error())
	}
	resourcesFS = fsys
}

type Game2 struct {
	mode Mode

	frameNum  uint64
	pressedA  bool
	releasedA bool
	pressedB  bool

	running  bool
	gopherY  fixed.Int26_6
	speedX   fixed.Int26_6
	speedY   fixed.Int26_6
	floating bool
	risingN  int

	bgMap  []uint8
	bgOffX fixed.Int26_6
	bgOffY fixed.Int26_6

	spPatterns []SpritePattern
	sprites    []Sprite

	groundHeight int
	groundHole   bool
	groundCont   int

	animeIndex int
	animeFrame int

	rand *rand.Rand

	bgTile     *ebiten.Image
	spriteTile *ebiten.Image

	seJump *audio.Player
}

func (g *Game2) Init() error {
	// Init BG
	g.bgMap = make([]uint8, scw*sch)

	// Init sprites
	g.spPatterns = []SpritePattern{
		{x: 0, y: 0, w: 16, h: 32},
		{x: 16, y: 0, w: 16, h: 32},
		{x: 32, y: 0, w: 16, h: 32},
		{x: 48, y: 0, w: 16, h: 32},
		{x: 64, y: 0, w: 16, h: 32},
		{x: 80, y: 0, w: 16, h: 32},
		{x: 96, y: 0, w: 16, h: 32},
	}
	g.sprites = []Sprite{
		{id: 0, x: int32(gopherX.Floor()), y: 0},
	}

	var err error
	g.bgTile, _, err = ebitenutil.NewImageFromFileSystem(resourcesFS, "chartable.png")
	if err != nil {
		return err
	}
	g.spriteTile, _, err = ebitenutil.NewImageFromFileSystem(resourcesFS, "spritetable.png")
	if err != nil {
		return err
	}

	// Load sounds
	audioContext := audio.NewContext(sampleRate)
	f, err := resourcesFS.Open("jump07.ogg")
	if err != nil {
		return err
	}
	defer f.Close()
	s, err := vorbis.DecodeWithSampleRate(sampleRate, f)
	if err != nil {
		return err
	}
	p, err := audioContext.NewPlayer(s)
	if err != nil {
		return err
	}
	g.seJump = p

	g.gotoTitle()
	return nil
}

func (g *Game2) gotoTitle() {
	for x := 0; x < scw; x++ {
		for y := 0; y < 10; y++ {
			g.bgMap[x*sch+y] = 0x00
		}
		for y := 10; y < sch; y++ {
			g.bgMap[x*sch+y] = 0x10
		}
	}
	g.mode = title
	g.running = true
	g.gopherY = gopherInitY
	g.speedX = initSpeedX
	g.speedY = 0
	g.floating = false
	g.risingN = 0
	g.bgOffX = 0

	g.groundHeight = 10
	g.groundHole = false
	g.groundCont = 5

	g.animeIndex = 0
	g.animeFrame = 0

	g.rand = rand.New(rand.NewSource(114514))
}

func (g *Game2) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 180
}

func isKeysJustPressed(keys ...ebiten.Key) bool {
	for _, k := range keys {
		if inpututil.IsKeyJustPressed(k) {
			return true
		}
	}
	return false
}

func isKeysJustReleased(keys ...ebiten.Key) bool {
	for _, k := range keys {
		if inpututil.IsKeyJustReleased(k) {
			return true
		}
	}
	return false
}

func (g *Game2) updateInput() error {
	g.running = !inpututil.IsKeyJustPressed(ebiten.KeyEscape)
	g.frameNum++
	g.pressedA = isKeysJustPressed(ebiten.KeyEnter, ebiten.KeySpace)
	g.releasedA = isKeysJustReleased(ebiten.KeyEnter, ebiten.KeySpace)
	g.pressedB = isKeysJustReleased(ebiten.KeyShift, ebiten.KeyControl)
	if !g.running {
		return errors.New("game aborted")
	}
	return nil
}

func (g *Game2) updateByInput() {
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
			g.seJump.SetPosition(0)
			g.seJump.Play()
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
}

func (g *Game2) checkToHitWalls() {
	if g.speedX <= 0 {
		return
	}
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

func (g *Game2) shiftBG() {
	l := len(g.bgMap)
	copy(g.bgMap[0:l-sch], g.bgMap[sch:])
}

// updateScroll scroll and prepare new area
func (g *Game2) updateScroll() {
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
			// FIXME: generate better stage data
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
}

func (g *Game2) checkToTouchGround() {
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
}

func (g *Game2) Update() error {
	if err := g.updateInput(); err != nil {
		return err
	}

	g.updateByInput()
	g.checkToHitWalls()
	g.updateScroll()
	g.gopherY += g.speedY
	g.checkToTouchGround()

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

	return nil
}

func (g *Game2) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Hello, World!")
	g.drawBG(screen)
	g.drawSprites(screen)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
}

func (g *Game2) drawBG(screen *ebiten.Image) {
	i := 0
	for x := 0; x < scw; x++ {
		dx := x*cellWidth - g.bgOffX.Floor()
		for y := 0; y < sch; y++ {
			dy := y*cellHeight - g.bgOffY.Floor()
			n := int(g.bgMap[i])
			sx := (n % 16) * cellWidth
			sy := (n / 16) * cellWidth
			cell := g.bgTile.SubImage(image.Rect(sx, sy, sx+cellWidth, sy+cellWidth)).(*ebiten.Image)
			var op ebiten.DrawImageOptions
			op.GeoM.Translate(float64(dx), float64(dy))
			screen.DrawImage(cell, &op)
			i++
		}
	}
}

func (g *Game2) drawSprites(screen *ebiten.Image) {
	for i := len(g.sprites) - 1; i >= 0; i-- {
		s := g.sprites[i]
		p := g.spPatterns[s.id]
		cell := g.spriteTile.SubImage(p.Rect()).(*ebiten.Image)
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(float64(s.x), float64(s.y))
		screen.DrawImage(cell, &op)
	}
}
