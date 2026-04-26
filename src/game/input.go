package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var StandardGamepadButtonMapping = map[ebiten.StandardGamepadButton]string{
	ebiten.StandardGamepadButtonRightBottom:      "RightBottom",
	ebiten.StandardGamepadButtonRightRight:       "RightRight",
	ebiten.StandardGamepadButtonRightLeft:        "RightLeft",
	ebiten.StandardGamepadButtonRightTop:         "RightTop",
	ebiten.StandardGamepadButtonFrontTopLeft:     "FrontTopLeft",
	ebiten.StandardGamepadButtonFrontTopRight:    "FrontTopRight",
	ebiten.StandardGamepadButtonFrontBottomLeft:  "FrontBottomLeft",
	ebiten.StandardGamepadButtonFrontBottomRight: "FrontBottomRight",
	ebiten.StandardGamepadButtonCenterLeft:       "CenterLeft",
	ebiten.StandardGamepadButtonCenterRight:      "CenterRight",
	ebiten.StandardGamepadButtonLeftStick:        "LeftStick",
	ebiten.StandardGamepadButtonRightStick:       "RightStick",
	ebiten.StandardGamepadButtonLeftTop:          "LeftTop",
	ebiten.StandardGamepadButtonLeftBottom:       "LeftBottom",
	ebiten.StandardGamepadButtonLeftLeft:         "LeftLeft",
	ebiten.StandardGamepadButtonLeftRight:        "LeftRight",
	ebiten.StandardGamepadButtonCenterCenter:     "CenterCenter",
}

func MarshalStandardGamepadButton(btn ebiten.StandardGamepadButton) string {
	if name, ok := StandardGamepadButtonMapping[btn]; ok {
		return name
	}
	return "Unknown"
}
func UnmarshalStandardGamepadButton(name string) (ebiten.StandardGamepadButton, bool) {
	for btn, btnName := range StandardGamepadButtonMapping {
		if btnName == name {
			return btn, true
		}
	}
	return 0, false
}

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

func (i *InputFrame) ProcessSmartInput(code string, playername string, playerConfigs map[int]*PlayerConfig) {
	// code is a string of 5 bits: [up, down, left, right, use_item]
	if len(code) != 5 {
		return
	}
	pid := -1
	for id, cfg := range playerConfigs {
		if cfg.Name == playername {
			pid = id
			break
		}
	}
	if pid == -1 {
		return
	}
	if i.Directions == nil {
		i.Directions = make(map[int]PlayerActionTurn)
	}
	if i.ItemsUsed == nil {
		i.ItemsUsed = make(map[int]bool)
	}
	for idx, char := range code {
		if char == '1' {
			switch idx {
			case 0:
				i.Directions[pid] = ActionLeft
			case 1:
				i.Directions[pid] = ActionRight
			case 2:
				i.Directions[pid] = ActionUp
			case 3:
				i.Directions[pid] = ActionDown
			case 4:
				i.ItemsUsed[pid] = true
			}
		}
	}
}

func (i *InputFrame) ProcessJustReleased(playerConfigs map[int]*PlayerConfig) {
	i.process(playerConfigs, inpututil.AppendJustReleasedKeys, inpututil.IsStandardGamepadButtonJustReleased)
}

func (i *InputFrame) Process(playerConfigs map[int]*PlayerConfig) {
	i.process(playerConfigs, inpututil.AppendJustPressedKeys, inpututil.IsStandardGamepadButtonJustPressed)
}

// Process reads current hardware inputs and populates Directions and ItemsUsed based on player keymaps.
func (i *InputFrame) process(playerConfigs map[int]*PlayerConfig, appendPressedKeys func([]ebiten.Key) []ebiten.Key, isGamepadButtonPressed func(ebiten.GamepadID, ebiten.StandardGamepadButton) bool) {
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

	//keys := inpututil.AppendJustPressedKeys(nil)
	keys := appendPressedKeys(nil)

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

		for _, gamepadid := range ebiten.AppendGamepadIDs([]ebiten.GamepadID{}) {
			if config.ControllerSDLID == ebiten.GamepadSDLID(gamepadid) {
				for actionStr, btnName := range config.ControllerMap {
					btn, ok := UnmarshalStandardGamepadButton(btnName)
					if !ok {
						continue
					}
					// inpututil.IsStandardGamepadButtonJustPressed(gamepadid, btn)
					if isGamepadButtonPressed(gamepadid, btn) {
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
}

var GlobalSmartController *SmartController

func InitSmartController(onNewConnection func()) {
	GlobalSmartController = &SmartController{
		OnNewConnection:  onNewConnection,
		ChanAssignPlayer: make(chan string),
	}
	go GlobalSmartController.StartServer(GConfig.SmartControllerPort)
}

func DefaultInputProcessor(playerConfigs map[int]*PlayerConfig) *InputFrame {
	input := &InputFrame{}
	input.Process(playerConfigs)
	if GlobalSmartController != nil {
		GlobalSmartController.ProcessInputs(input, playerConfigs)
	}
	return input
}
