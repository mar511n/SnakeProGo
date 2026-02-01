package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	menu *MainMenu
}

func (g *Game) Update() error {
	if g.menu == nil {
		g.menu = NewMainMenu()
	}
	return g.menu.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.menu != nil {
		g.menu.Draw(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("SnakeProGo")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
