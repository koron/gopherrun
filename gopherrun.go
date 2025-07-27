package main

import (
	_ "embed"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	var game Game
	if err := game.Init(); err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("Gopher Run!")
	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
