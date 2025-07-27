package main

import (
	_ "embed"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
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
