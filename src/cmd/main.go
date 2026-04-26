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
	//ModeReplay
)

type Game struct {
	mode    GameMode
	menu    *game.MainMenu
	session *game.GameSession
	//replay    *game.ReplaySession
	renderer  game.Renderer
	resources *game.ResourceManager
}

func (g *Game) Update() error {
	if g.mode == ModePlaying {
		g.session.Update()
		return nil
		/*
			} else if g.mode == ModeControllerConfig {
			anythingpressed := false
			pressedBtns := make(map[string][]string)
			for _, id := range ebiten.AppendGamepadIDs([]ebiten.GamepadID{}) {
				btns := inpututil.AppendJustPressedStandardGamepadButtons(id, []ebiten.StandardGamepadButton{})
				pressedBtns[ebiten.GamepadSDLID(id)] = make([]string, len(btns))
				for i, btn := range btns {
					pressedBtns[ebiten.GamepadSDLID(id)][i] = game.MarshalStandardGamepadButton(btn)
				}
				if len(pressedBtns[ebiten.GamepadSDLID(id)]) > 0 {
					anythingpressed = true
				}
			}
			if anythingpressed {
				g.menu.AddHistory("Pressed buttons: %v", pressedBtns)
				game.LogInfo("Pressed buttons: %v", pressedBtns)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
				g.mode = ModeMenu
			}
			g.menu.Update()
			return nil
		*/
	} else {
		return g.menu.Update()
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.mode == ModePlaying {
		g.renderer.Render(g.session.State, screen)
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
	if _, err := os.Stat(game.BaseSystemPath); os.IsNotExist(err) {
		game.LogInfo("Base path %s does not exist, creating it...", game.BaseSystemPath)
		if err := os.MkdirAll(game.BaseSystemPath, 0755); err != nil {
			game.FatalError("Failed to create base path %s: %v", game.BaseSystemPath, err)
		}
	}
	game.LogInfo("Using base path: %s", game.BaseSystemPath)

	game.LoadConfigs()
	game.InitLLM()

	ebitengame := &Game{mode: ModeMenu}
	ebitengame.resources = &game.ResourceManager{}
	ebitengame.resources.LoadAssets(filepath.Join(game.BaseSystemPath, game.ResDir, game.AssetsDir), "Images", "Sounds", "Icons", "Font")

	ebiten.SetWindowIcon(ebitengame.resources.Icons)
	game.InitSound(ebitengame.resources)

	var OnGameOver = func(winnerIDs []int, hist *game.HistoryData) {
		winnernames := make([]string, len(winnerIDs))
		for i, id := range winnerIDs {
			if player, ok := ebitengame.session.State.Players[id]; ok {
				winnernames[i] = player.Config.Name
			} else {
				winnernames[i] = "Unknown"
			}
		}
		ebitengame.menu.PrintMessage("Game over! Winners: %v", winnernames)

		ebitengame.menu.OnGameSessionDone(winnernames, hist, ebitengame.resources)

		ebitengame.mode = ModeMenu
	}
	var OnStartGame = func() {
		ebitengame.session = game.NewGameSession(OnGameOver)
		ebitengame.session.Initialize()
		ebitengame.renderer = game.NewDefaultRenderer(ebitengame.resources)
		ebitengame.renderer.InitRender(true, ebitengame.session.State)
		ebitengame.mode = ModePlaying
	}

	ebitengame.menu = game.NewMainMenu(OnStartGame, ebitengame.resources)
	/*
		ebitengame.menu.OnUnknownCommand = func(cmd string) bool {
			if game.PConfigs != nil {
				game.LogInfo("Received unknown command '%s', checking if it matches any player config names for smart controller assignment...", cmd)

				game.LogInfo("Registered player configs: %v", game.PConfigs)
				for name, cfg := range game.PConfigs {
					if cmd == cfg.Name {
						if game.GlobalSmartController != nil {
							game.LogInfo("Assigning smart controller to player '%s'", name)
							game.GlobalSmartController.ChanAssignPlayer <- name
							ebitengame.menu.AddHistory("Assigned smart controller to player '%s'", name)
							return true
						}
					}
				}
			}
			return false
		}
	*/
	/*
		game.InitSmartController(func() {
			if ebitengame.mode == ModeControllerConfig {
				ebitengame.menu.AddHistory("New smartphone controller connected. Choose a player for it. (Enter player name as command)")
			}
		})
	*/

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
