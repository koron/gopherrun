package main

import (
	"image"

	"golang.org/x/image/math/fixed"
)

const (
	screenWidth  = 320
	screenHeight = 180
	cellWidth    = 16
	cellHeight   = 16

	sampleRate = 44100

	// scw is screen character width
	scw = (screenWidth+cellWidth-1)/cellWidth + 1

	// sch is screen character height
	sch = (screenHeight+cellHeight-1)/cellHeight + 1

	risingInitN = 10
)

var (
	// maxBgOffx is max value for bgOffX
	maxBgOffx = fixed.I(16)

	// gravityPower
	gravityPower = fixed.I(3) / 2

	maxSpeedY = fixed.I(16) / 2

	// risingPower
	risingPower = fixed.I(11) / 2

	gopherX = fixed.I(320) / 5

	initSpeedX = fixed.I(1) / 3

	maxSpeedX = fixed.I(6) / 2

	accelX = fixed.I(1) / 40

	gopherInitY = fixed.I(8 * 16)

	walkPattern = []int{3, 4, 5}
)

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

func (sp SpritePattern) Rect() image.Rectangle {
	return image.Rect(int(sp.x), int(sp.y), int(sp.x+sp.w), int(sp.y+sp.h))
}

type Mode int

const (
	title Mode = iota
	playing
	gameover
)
