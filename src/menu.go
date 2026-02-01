package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type MainMenu struct {
	inputBuffer []rune
	history     []string
	cursorBlink int
	players     []string
}

func NewMainMenu() *MainMenu {
	return &MainMenu{
		history: []string{},
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
	header := "Welcome to SnakeProGo!\n" +
		"Available commands:\n" +
		"  addplayer <name>\n" +
		"  removeplayer <name>\n" +
		"  listplayers\n" +
		"  startgame\n" +
		"  quit\n" +
		"----------------------------------------\n"

	// Show last N lines of history + current input
	const maxLines = 15

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
		if len(args) < 1 {
			m.history = append(m.history, "Usage: addplayer <name>")
		} else {
			name := args[0]
			m.players = append(m.players, name)
			m.history = append(m.history, fmt.Sprintf("Added player: %s", name))
		}
	case "removeplayer":
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
				m.history = append(m.history, fmt.Sprintf("Removed player: %s", name))
			} else {
				m.history = append(m.history, fmt.Sprintf("Player not found: %s", name))
			}
		}
	case "listplayers":
		if len(m.players) == 0 {
			m.history = append(m.history, "No players joined.")
		} else {
			m.history = append(m.history, "Current players: "+strings.Join(m.players, ", "))
		}
	case "startgame":
		m.history = append(m.history, "Starting game... (not implemented)")
		// TODO: Transition to game state
	case "quit":
		m.history = append(m.history, "Quitting...")
		os.Exit(0)
	default:
		m.history = append(m.history, fmt.Sprintf("Unknown command: %s", command))
	}
}
