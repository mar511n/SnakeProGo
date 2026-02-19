package game

import (
	"fmt"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/pelletier/go-toml/v2"
)

type MainMenu struct {
	inputBuffer        []rune
	history            []string
	cursorBlink        int
	players            []string
	OnStartGame        func()
	OnReplay           func()
	OnControllerConfig func()
	OnUnknownCommand   func(cmd string) bool
}

func NewMainMenu(startGameCallback func(), controllerConfigCallback func()) *MainMenu {
	return &MainMenu{
		history: []string{
			"Welcome to SnakeProGo!",
		},
		players:            []string{},
		OnStartGame:        startGameCallback,
		OnControllerConfig: controllerConfigCallback,
	}
}

func (m *MainMenu) AddHistory(format string, args ...interface{}) {
	entry := fmt.Sprintf(format, args...)
	m.history = append(m.history, entry)
}

func repeatingKeyPressed(key ebiten.Key) bool {
	const (
		delay    = 30
		interval = 3
	)
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true
	}
	if d >= delay && (d-delay)%interval == 0 {
		return true
	}
	return false
}

func (m *MainMenu) Update() error {
	// Handle character input
	m.inputBuffer = ebiten.AppendInputChars(m.inputBuffer)

	// Handle Backspace
	if repeatingKeyPressed(ebiten.KeyBackspace) {
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}
	}

	// Handle Enter
	if repeatingKeyPressed(ebiten.KeyEnter) || repeatingKeyPressed(ebiten.KeyKPEnter) {
		cmdStr := string(m.inputBuffer)
		m.AddHistory("> " + cmdStr)
		m.processCommand(cmdStr)
		m.inputBuffer = m.inputBuffer[:0]
	}

	m.cursorBlink++
	return nil
}

func (m *MainMenu) Draw(screen *ebiten.Image) {
	header := "Available commands:\n" +
		"  addplayer (apl) <name>\n" +
		"  removeplayer (rpl) <name>\n" +
		"  listplayers (lpl)\n" +
		"  replay (r)\n" +
		"  showconfig (sc) [global/game/username]\n" +
		"  controller (c)\n" +
		"  startgame (start)\n" +
		"  quit (q)\n" +
		"----------------------------------------\n"

	// Show last N lines of history + current input
	var maxLines = GConfig.ScreenHeight / 16

	start := 0
	if len(m.history) > maxLines-11 {
		start = len(m.history) - maxLines + 10 + 1
	}

	displayText := header + strings.Join(m.history[start:], "\n")
	displayText += "\n> " + string(m.inputBuffer)

	if m.cursorBlink%60 < 30 {
		displayText += "_"
	}

	ebitenutil.DebugPrint(screen, displayText)
}

func (m *MainMenu) processCommand(cmd string) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case "controller", "c":
		m.AddHistory("Entering controller configuration mode. Press any button on your controller to see its Name and SDL ID in the history. Press Escape to exit this mode.")
		ids := ebiten.AppendGamepadIDs([]ebiten.GamepadID{})
		sdlids := make([]string, len(ids))
		for i, id := range ids {
			sdlids[i] = ebiten.GamepadSDLID(id)
		}
		m.AddHistory("Detected controllers: %v", sdlids)
		if m.OnControllerConfig != nil {
			m.OnControllerConfig()
		}
	case "replay", "r":
		m.AddHistory("Replaying game...")
		if m.OnReplay != nil {
			m.OnReplay()
		}
	case "addplayer", "apl":
		if len(args) < 1 {
			m.AddHistory("Usage: addplayer <name>")
		} else {
			name := args[0]
			// Check if already added to list
			alreadyAdded := false
			for _, p := range m.players {
				if p == name {
					alreadyAdded = true
					break
				}
			}

			if alreadyAdded {
				m.AddHistory("Player %s is already in the game.", name)
			} else {
				m.players = append(m.players, name)
				GetPlayerConfig(name) // Ensure config is loaded/created
				m.AddHistory("Added player: %s", name)
			}
		}
	case "showconfig", "sc":
		if len(args) < 1 {
			m.AddHistory("Usage: showconfig [global/game/<username>]")
		} else {
			target := args[0]
			var cfg interface{}
			found := true

			switch target {
			case "global":
				cfg = GConfig
			case "game":
				cfg = GPConfig
			default:
				// Assume player name
				if _, ok := PConfigs[target]; ok {
					cfg = PConfigs[target]
				} else {
					found = false
					m.AddHistory("Config not found for: %s. (Is player added?)", target)
				}
			}

			if found {
				data, err := toml.Marshal(cfg)
				if err != nil {
					m.AddHistory("Error displaying config: %v", err)
				} else {
					// Split lines so they show up nicely in history
					lines := strings.Split(string(data), "\n")
					m.AddHistory("--- Config: %s ---", target)
					m.history = append(m.history, lines...)
					m.AddHistory("----------------")
				}
			}
		}
	case "removeplayer", "rpl":
		if len(args) < 1 {
			m.AddHistory("Usage: removeplayer <name>")
		} else {
			name := args[0]
			found := false
			newPlayers := []string{}
			for _, p := range m.players {
				if p == name {
					found = true
				} else {
					newPlayers = append(newPlayers, p)
				}
			}
			if found {
				m.players = newPlayers
				delete(PConfigs, name)
				m.AddHistory("Removed player: %s", name)
			} else {
				m.AddHistory("Player not found: %s", name)
			}
		}
	case "listplayers", "lpl":
		if len(m.players) == 0 {
			m.AddHistory("No players joined.")
		} else {
			m.AddHistory("Current players: " + strings.Join(m.players, ", "))
		}
	case "startgame", "start":
		m.AddHistory("Starting game...")
		if m.OnStartGame != nil {
			m.OnStartGame()
		}
	case "quit", "exit", "q":
		m.AddHistory("Quitting...")
		os.Exit(0)
	default:
		if m.OnUnknownCommand != nil {
			handled := m.OnUnknownCommand(command)
			if handled {
				return
			} else {
				m.AddHistory("Unknown command: %s", command)
			}
		}
	}
}
