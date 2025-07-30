package main

import (
	"embed"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"io/fs"
	"iter"
	"log/slog"
	"math/rand"
	"sync"

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

type Game struct {
	frameNum uint64
	rand     *rand.Rand

	updateNext func() (error, bool)
	updateStop func()

	jumpPressed  bool
	jumpReleased bool

	groundHeight int
	groundHole   bool
	groundCont   int

	gopherY  fixed.Int26_6
	speedX   fixed.Int26_6
	speedY   fixed.Int26_6
	floating bool
	risingN  int

	animeIndex int
	animeFrame int

	seJump *audio.Player

	Screen struct {
		bgTile *ebiten.Image
		bgMap  []uint8
		bgOffX fixed.Int26_6
		bgOffY fixed.Int26_6

		spriteTile *ebiten.Image
		spPatterns []SpritePattern
		sprites    []Sprite
	}
}

func (g *Game) Init() error {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		g.updateNext, g.updateStop = iter.Pull(g.yieldUpdate)
		wg.Done()
	}()

	// Init BG
	g.Screen.bgMap = make([]uint8, scw*sch)

	// Init sprites
	g.Screen.spPatterns = []SpritePattern{
		{x: 0, y: 0, w: 16, h: 32},
		{x: 16, y: 0, w: 16, h: 32},
		{x: 32, y: 0, w: 16, h: 32},
		{x: 48, y: 0, w: 16, h: 32},
		{x: 64, y: 0, w: 16, h: 32},
		{x: 80, y: 0, w: 16, h: 32},
		{x: 96, y: 0, w: 16, h: 32},
	}
	g.Screen.sprites = []Sprite{
		{id: 0, x: int(gopherX.Floor()), y: 0},
	}

	var err error
	g.Screen.bgTile, _, err = ebitenutil.NewImageFromFileSystem(resourcesFS, "chartable.png")
	if err != nil {
		return err
	}
	g.Screen.spriteTile, _, err = ebitenutil.NewImageFromFileSystem(resourcesFS, "spritetable.png")
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

	wg.Wait()
	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 180
}

var ErrGameAborted = errors.New("game aborted")

func (g *Game) yieldInput(yield func(error) bool) bool {
	g.frameNum++
	g.jumpPressed = isKeysJustPressed(ebiten.KeyEnter, ebiten.KeySpace)
	g.jumpReleased = isKeysJustReleased(ebiten.KeyEnter, ebiten.KeySpace)
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		yield(ErrGameAborted)
		return false
	}
	return true
}

func (g *Game) yieldUpdate(yield func(error) bool) {
	for {
		if !g.yieldTitle(yield) {
			break
		}
		if !g.yieldPlaying(yield) {
			break
		}
		if !g.yieldGameover(yield) {
			break
		}
	}
}

func (g *Game) yieldTitle(yield func(error) bool) bool {
	slog.Info("Mode: Title")

	// Setup title
	for x := 0; x < scw; x++ {
		for y := 0; y < 10; y++ {
			g.Screen.bgMap[x*sch+y] = 0x00
		}
		for y := 10; y < sch; y++ {
			g.Screen.bgMap[x*sch+y] = 0x10
		}
	}
	g.gopherY = gopherInitY
	g.speedX = initSpeedX
	g.speedY = 0
	g.floating = false
	g.risingN = 0
	g.Screen.bgOffX = 0
	g.groundHeight = 10
	g.groundHole = false
	g.groundCont = 5
	g.animeIndex = 0
	g.animeFrame = 0
	g.rand = rand.New(rand.NewSource(114514))

	g.Screen.sprites[0].y = int(gopherInitY.Floor())

	for g.yieldInput(yield) {
		if g.jumpPressed {
			// Exit the title
			return true
		}

		g.Screen.bgOffX += initSpeedX

		for g.Screen.bgOffX >= maxBgOffx {
			g.Screen.bgOffX -= maxBgOffx
			g.shiftBG()
			// insert new bgMap at right
			n := (scw - 1) * sch
			for y := 0; y < sch; y++ {
				if y < 10 {
					g.Screen.bgMap[n+y] = 0x00
				} else {
					g.Screen.bgMap[n+y] = 0x10
				}
			}
		}

		g.updateGopher()
		g.Screen.sprites[0].id = g.animeIndex

		if !yield(nil) {
			return false
		}
	}
	return false
}

func (g *Game) yieldPlaying(yield func(error) bool) bool {
	slog.Info("Mode: Playing")
	for g.yieldInput(yield) {
		// Update the gopher state.
		if g.floating {
			if g.risingN > 0 {
				if g.jumpReleased {
					g.risingN = 0
				} else {
					g.risingN--
				}
				g.speedY = -risingPower
			} else {
				g.speedY = min(g.speedY+gravityPower, maxSpeedY)
			}
		} else {
			if g.jumpPressed {
				g.floating = true
				g.risingN = risingInitN
				g.speedY = -risingPower
				g.seJump.SetPosition(0)
				g.seJump.Play()
			}
		}
		g.speedX = min(g.speedX+accelX, maxSpeedX)
		g.Screen.bgOffX += g.speedX
		g.checkToHitWalls()

		// Update scroll
		for g.Screen.bgOffX >= maxBgOffx {
			g.Screen.bgOffX -= maxBgOffx
			g.shiftBG()
			// Insert new bgMap at right
			n := (scw - 1) * sch
			for y := range sch {
				if !g.groundHole && y >= g.groundHeight {
					g.Screen.bgMap[n+y] = 0x10
				} else {
					g.Screen.bgMap[n+y] = 0x00
				}
			}
			// FIXME: generate better stage data
			g.groundCont--
			if g.groundCont <= 0 {
				g.generateNextBlocks()
			}
		}
		g.gopherY += g.speedY
		g.checkToTouchGround()

		// Check gameover
		if g.gopherY.Floor() > screenHeight {
			// FIXME: show game over message
			yield(nil)
			return true
		}

		if g.floating {
			g.animeIndex = 6
			g.animeFrame = 0
		} else if g.speedX > 0 {
			g.updateGopher()
		} else {
			g.animeFrame = 0
			g.animeIndex = 0
		}

		g.Screen.sprites[0].y = int(g.gopherY.Floor())
		g.Screen.sprites[0].id = g.animeIndex
		if !yield(nil) {
			return false
		}
	}
	return false
}

func (g *Game) yieldGameover(yield func(error) bool) bool {
	slog.Info("Mode: Gameover")
	g.speedX = 0
	for g.yieldInput(yield) {
		if g.jumpPressed {
			// Exit gameover, and go to title.
			return yield(nil)
		}
		if !yield(nil) {
			return false
		}
	}
	return false
}

func (g *Game) shiftBG() {
	l := len(g.Screen.bgMap)
	copy(g.Screen.bgMap[0:l-sch], g.Screen.bgMap[sch:])
}

func (g *Game) checkToHitWalls() {
	if g.speedX <= 0 {
		return
	}
	y := g.gopherY.Floor()
	cx := ((gopherX + g.Screen.bgOffX).Floor() + cellWidth) / cellWidth
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
		if g.Screen.bgMap[cx*sch+cy+i] >= 0x10 {
			hit = true
			break
		}
	}
	if hit {
		g.speedX = 0
		g.Screen.bgOffX = fixed.I((cx-1)*cellWidth) - gopherX
	}
}

func (g *Game) generateNextBlocks() {
	if !g.groundHole && g.rand.Float32() < 0.17 {
		c := between(int(g.rand.ExpFloat64()*1.5), 1, 4)
		g.groundHole = true
		g.groundCont = c
		return
	}

	if r := g.rand.Float32(); r < 0.18 {
		c := between(int(g.rand.ExpFloat64()*1), 1, 4)
		g.groundHeight = max(g.groundHeight-c, 4)
		if g.groundHeight < 4 {
			g.groundHeight = 4
		}
	} else if r >= 0.82 {
		c := between(int(g.rand.ExpFloat64()*1), 1, 4)
		g.groundHeight = min(g.groundHeight+c, 10)
	}
	g.groundHole = false
	g.groundCont = between(int(g.rand.NormFloat64()*2+3), 1, 8)
}

func (g *Game) checkToTouchGround() {
	// check to touch grand
	if g.speedY >= 0 {
		x := (gopherX + g.Screen.bgOffX).Floor()
		cx := x / cellWidth
		cy := (g.gopherY.Floor() + cellHeight*2) / cellHeight
		cw := 2
		if x%cellWidth == 0 {
			cw = 1
		}
		if cy >= 0 && cy < sch {
			touch := false
			for i := 0; i < cw; i++ {
				if g.Screen.bgMap[(cx+i)*sch+cy] >= 0x10 {
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
}

func (g *Game) updateGopher() {
	g.animeIndex = walkPattern[g.animeFrame/10]
	g.animeFrame++
	if g.animeFrame >= len(walkPattern)*10 {
		g.animeFrame = 0
	}
}

func (g *Game) Update() error {
	err, ok := g.updateNext()
	if err != nil || !ok {
		g.updateStop()
		return err
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Hello, World!")
	g.drawBG(screen)
	g.drawSprites(screen)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))
}

func (g *Game) drawBG(screen *ebiten.Image) {
	i := 0
	for x := 0; x < scw; x++ {
		dx := x*cellWidth - g.Screen.bgOffX.Floor()
		for y := 0; y < sch; y++ {
			dy := y*cellHeight - g.Screen.bgOffY.Floor()
			n := int(g.Screen.bgMap[i])
			sx := (n % 16) * cellWidth
			sy := (n / 16) * cellWidth
			cell := g.Screen.bgTile.SubImage(image.Rect(sx, sy, sx+cellWidth, sy+cellWidth)).(*ebiten.Image)
			var op ebiten.DrawImageOptions
			op.GeoM.Translate(float64(dx), float64(dy))
			screen.DrawImage(cell, &op)
			i++
		}
	}
}

func (g *Game) drawSprites(screen *ebiten.Image) {
	for i := len(g.Screen.sprites) - 1; i >= 0; i-- {
		s := g.Screen.sprites[i]
		p := g.Screen.spPatterns[s.id]
		cell := g.Screen.spriteTile.SubImage(p.Rect()).(*ebiten.Image)
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(float64(s.x), float64(s.y))
		screen.DrawImage(cell, &op)
	}
}
