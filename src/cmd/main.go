package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"

	"SnakeProGo/game"
)

type Game struct {
	menu *game.MainMenu
}

func (g *Game) Update() error {
	if g.menu == nil {
		g.menu = game.NewMainMenu()
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
	game.LogInfo("Starting SnakeProGo...")
	userHome, err := os.UserHomeDir()
	if err != nil {
		userHome = "." // Fallback to current directory if home cannot be determined
	}
	defaultPath := filepath.Join(userHome, "snakeprogo")

	flag.StringVar(&game.BaseSystemPath, "base-path", defaultPath, "Base path for game data")
	flag.Parse()

	// Ensure base system path exists
	if _, err := os.Stat(game.BaseSystemPath); os.IsNotExist(err) {
		game.LogInfo("Base path %s does not exist, creating it...", game.BaseSystemPath)
		if err := os.MkdirAll(game.BaseSystemPath, 0755); err != nil {
			game.FatalError("Failed to create base path %s: %v", game.BaseSystemPath, err)
		}
	}

	game.LoadConfigs()
	// Initialize default player configs to show in list if needed, or wait for addplayer

	ebiten.SetWindowTitle("SnakeProGo")
	if err := ebiten.RunGame(&Game{}); err != nil {
		game.FatalError("Game crashed: %v", err)
	}
}
