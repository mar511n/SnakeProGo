package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"

	"SnakeProGo/game"
)

type GameMode int

const (
	ModeMenu GameMode = iota
	ModePlaying
	ModeReplay
)

type Game struct {
	mode      GameMode
	menu      *game.MainMenu
	session   *game.GameSession
	replay    *game.ReplaySession
	renderer  game.Renderer
	resources *game.ResourceManager
}

func (g *Game) Update() error {
	if g.mode == ModePlaying {
		g.session.Update()
		return nil
	} else if g.mode == ModeReplay {
		if g.replay.IsFinished() {
			g.mode = ModeMenu
			g.menu.AddHistory("Replay finished.")
			g.replay = game.NewReplaySession(g.replay.History)
		} else {
			g.replay.Update()
		}
		return nil
	} else {
		return g.menu.Update()
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.mode == ModePlaying {
		g.renderer.Render(g.session.State, screen)
	} else if g.mode == ModeReplay {
		g.renderer.Render(g.replay.State, screen)
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
	ebitengame := &Game{mode: ModeMenu}
	ebitengame.menu = game.NewMainMenu(func() {
		ebitengame.session = game.NewGameSession(func(winnerIDs []int, hist *game.HistoryData) {
			ebitengame.mode = ModeMenu
			winnernames := make([]string, len(winnerIDs))
			for i, id := range winnerIDs {
				if player, ok := ebitengame.session.State.Players[id]; ok {
					winnernames[i] = player.Config.Name
				} else {
					winnernames[i] = "Unknown"
				}
			}
			ebitengame.menu.AddHistory("Game over! Winners: %v", winnernames)
			ebitengame.replay = game.NewReplaySession(hist)
			ebitengame.menu.OnReplay = func() {
				ebitengame.mode = ModeReplay
			}
		})
		ebitengame.session.Initialize()
		ebitengame.renderer = game.NewDefaultRenderer(ebitengame.resources)
		ebitengame.renderer.InitRender(true, ebitengame.session.State)
		ebitengame.mode = ModePlaying
	})
	ebitengame.resources = &game.ResourceManager{}
	ebitengame.resources.LoadAssets(filepath.Join(game.BaseSystemPath, game.ResDir, game.AssetsDir), "Images", "Sounds", "Icons")
	ebiten.SetWindowIcon(ebitengame.resources.Icons)

	game.InitSound(ebitengame.resources)
	sound, ok := ebitengame.resources.RandomSound("Revive")
	if ok {
		sound.GetPlayer().Play()
	} else {
		game.LogWarning("No 'Revive' sound found to play on startup.")
	}

	if err := ebiten.RunGame(ebitengame); err != nil {
		game.FatalError("Game crashed: %v", err)
	}
}
