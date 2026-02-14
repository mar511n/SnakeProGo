package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type PlayerActionTurn int

const (
	ActionNone PlayerActionTurn = iota
	ActionUp
	ActionDown
	ActionLeft
	ActionRight
	ActionTurnLeft
	ActionTurnRight
)

func (a PlayerActionTurn) IsValid(facing Vec2i) bool {
	switch a {
	case ActionUp, ActionDown:
		return facing != DirDown && facing != DirUp
	case ActionLeft, ActionRight:
		return facing != DirLeft && facing != DirRight
	case ActionTurnLeft, ActionTurnRight:
		return true
	default:
		return false
	}
}

type InputFrame struct {
	Tick       uint64
	Directions map[int]PlayerActionTurn // Keyed by player ID
	ItemsUsed  map[int]bool             // Keyed by player ID, true indicates item usage
}

// Process reads current hardware inputs and populates Directions and ItemsUsed based on player keymaps.
func (i *InputFrame) Process(playerConfigs map[int]*PlayerConfig) {
	if i.Directions == nil {
		i.Directions = make(map[int]PlayerActionTurn)
	}
	if i.ItemsUsed == nil {
		i.ItemsUsed = make(map[int]bool)
	}

	actionMap := map[string]PlayerActionTurn{
		"up":         ActionUp,
		"down":       ActionDown,
		"left":       ActionLeft,
		"right":      ActionRight,
		"turn_left":  ActionTurnLeft,
		"turn_right": ActionTurnRight,
	}

	keys := inpututil.AppendJustPressedKeys(nil)

	for pID, config := range playerConfigs {
		i.Directions[pID] = ActionNone
		i.ItemsUsed[pID] = false

		for _, key := range keys {
			for actionStr, keyName := range config.KeyMap {
				var configuredKey ebiten.Key
				if err := configuredKey.UnmarshalText([]byte(keyName)); err != nil {
					continue
				}

				if key == configuredKey {
					if actionStr == "use_item" {
						i.ItemsUsed[pID] = true
					} else if act, ok := actionMap[actionStr]; ok {
						i.Directions[pID] = act
					}
				}
			}
		}
	}
}
