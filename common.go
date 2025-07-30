package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/math/fixed"
)

const (
	windowWidth  = screenWidth * 4
	windowHeight = screenHeight * 4

	screenWidth  = 320
	screenHeight = 180
	cellWidth    = 16
	cellHeight   = 16

	sampleRate = 44100

	// scw is screen character width
	scw = (screenWidth+cellWidth-1)/cellWidth + 1

	// sch is screen character height
	sch = (screenHeight+cellHeight-1)/cellHeight + 1

	// maxBgOffx is max value for bgOffX
	maxBgOffx = fixed.Int26_6(16 << 6)

	risingInitN = 10

	risingPower = fixed.Int26_6(11<<6) / 2

	gravityPower = fixed.Int26_6(3<<6) / 2

	maxSpeedY = fixed.Int26_6(16<<6) / 2

	initSpeedX = fixed.Int26_6(1<<6) / 3

	accelX = fixed.Int26_6(1<<6) / 40

	maxSpeedX = fixed.Int26_6(6<<6) / 2

	gopherX = fixed.Int26_6(320<<6) / 5

	gopherInitY = fixed.Int26_6(8 * 16 << 6)
)

var walkPattern = []int{3, 4, 5}

type Sprite struct {
	id int
	x  int
	y  int
}

type SpritePattern struct {
	x int
	y int
	w int
	h int
}

func (sp SpritePattern) Rect() image.Rectangle {
	return image.Rect(sp.x, sp.y, sp.x+sp.w, sp.y+sp.h)
}

func between(n, minN, maxN int) int {
	return min(max(n, minN), maxN)
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
