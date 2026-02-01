package main

import (
	"flag"
	"os"
	"path/filepath"

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
	LogInfo("Starting SnakeProGo...")
	userHome, err := os.UserHomeDir()
	if err != nil {
		userHome = "." // Fallback to current directory if home cannot be determined
	}
	defaultPath := filepath.Join(userHome, "snakeprogo")

	flag.StringVar(&BaseSystemPath, "base-path", defaultPath, "Base path for game data")
	flag.Parse()

	// Ensure base system path exists
	if _, err := os.Stat(BaseSystemPath); os.IsNotExist(err) {
		LogInfo("Base path %s does not exist, creating it...", BaseSystemPath)
		if err := os.MkdirAll(BaseSystemPath, 0755); err != nil {
			FatalError("Failed to create base path %s: %v", BaseSystemPath, err)
		}
	}

	LoadConfigs()
	// Initialize default player configs to show in list if needed, or wait for addplayer

	ebiten.SetWindowTitle("SnakeProGo")
	if err := ebiten.RunGame(&Game{}); err != nil {
		FatalError("Game crashed: %v", err)
	}
}
