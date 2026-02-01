package main

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
	inputBuffer []rune
	history     []string
	cursorBlink int
	players     []string
}

func NewMainMenu() *MainMenu {
	return &MainMenu{
		history: []string{
			"Welcome to SnakeProGo!",
		},
		players: []string{},
	}
}

// repeatingKeyPressed return true when key is pressed considering the repeat state.
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
		m.history = append(m.history, "> "+cmdStr)
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
		"  showconfig (sc) [global/game/username]\n" +
		"  startgame (start)\n" +
		"  quit (q)\n" +
		"----------------------------------------\n"

	// Show last N lines of history + current input
	const maxLines = 20

	start := 0
	if len(m.history) > maxLines {
		start = len(m.history) - maxLines
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
	case "addplayer":
	case "apl":
		if len(args) < 1 {
			m.history = append(m.history, "Usage: addplayer <name>")
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
				m.history = append(m.history, fmt.Sprintf("Player %s is already in the game.", name))
			} else {
				m.players = append(m.players, name)
				GetPlayerConfig(name) // Ensure config is loaded/created
				m.history = append(m.history, fmt.Sprintf("Added player: %s", name))
			}
		}
	case "showconfig":
	case "sc":
		if len(args) < 1 {
			m.history = append(m.history, "Usage: showconfig [global/game/<username>]")
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
					m.history = append(m.history, fmt.Sprintf("Config not found for: %s. (Is player added?)", target))
				}
			}

			if found {
				data, err := toml.Marshal(cfg)
				if err != nil {
					m.history = append(m.history, fmt.Sprintf("Error displaying config: %v", err))
				} else {
					// Split lines so they show up nicely in history
					lines := strings.Split(string(data), "\n")
					m.history = append(m.history, fmt.Sprintf("--- Config: %s ---", target))
					m.history = append(m.history, lines...)
					m.history = append(m.history, "----------------")
				}
			}
		}
	case "removeplayer":
	case "rpl":
		if len(args) < 1 {
			m.history = append(m.history, "Usage: removeplayer <name>")
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
				m.history = append(m.history, fmt.Sprintf("Removed player: %s", name))
			} else {
				m.history = append(m.history, fmt.Sprintf("Player not found: %s", name))
			}
		}
	case "listplayers":
	case "lpl":
		if len(m.players) == 0 {
			m.history = append(m.history, "No players joined.")
		} else {
			m.history = append(m.history, "Current players: "+strings.Join(m.players, ", "))
		}
	case "startgame":
	case "start":
		m.history = append(m.history, "Starting game... (not implemented)")
		// TODO: Transition to game state
	case "quit":
	case "exit":
	case "q":
		m.history = append(m.history, "Quitting...")
		os.Exit(0)
	default:
		m.history = append(m.history, fmt.Sprintf("Unknown command: %s", command))
	}
}
