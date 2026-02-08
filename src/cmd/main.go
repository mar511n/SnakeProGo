package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"

	"SnakeProGo/game"
)

type Game struct {
	is_running bool
	menu       *game.MainMenu
	session    *game.GameSession
}

func (g *Game) Update() error {
	if g.is_running {
		g.session.Update()
		return nil
	}
	return g.menu.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.is_running {
		game.Render(g.session.State, screen)
	} else {
		g.menu.Draw(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return game.GConfig.ScreenWidth, game.GConfig.ScreenHeight
}

func main() {
	ebiten.SetWindowTitle("SnakeProGo")
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
	game.LogInfo("Using base path: %s", game.BaseSystemPath)

	game.LoadConfigs()
	ebitengame := &Game{is_running: false}
	ebitengame.menu = game.NewMainMenu(func() {
		ebitengame.session = game.NewGameSession(func(winnerIDs []int) {
			ebitengame.is_running = false
			winnernames := make([]string, len(winnerIDs))
			for i, id := range winnerIDs {
				if player, ok := ebitengame.session.State.Players[id]; ok {
					winnernames[i] = player.Config.Name
				} else {
					winnernames[i] = "Unknown"
				}
			}
			ebitengame.menu.AddHistory("Game over! Winners: %v", winnernames)
		})
		ebitengame.session.Initialize()
		ebitengame.is_running = true
	})

	if err := ebiten.RunGame(ebitengame); err != nil {
		game.FatalError("Game crashed: %v", err)
	}
}
