package game

import (
	"os"
	"path/filepath"
)

type GameEvent struct {
	ID      uint64
	Type    string
	Payload interface{}
}

// GameState holds the complete state of the simulation at a specific tick.
type GameState struct {
	Tick     uint64
	Map      *MapData
	Players  map[int]*PlayerSnake // Keyed by player ID
	Apples   []*Apple             // Collectible apples
	Items    []*Item              // Collectible items
	Entities []Entity             // Dynamic entities (Bullets, Farts, Bots)
	Events   []*GameEvent         // for audio and visual effects
}

type GameSession struct {
	State             *GameState
	RegisteredPlayers map[int]*PlayerConfig // Key: ID, Value: Reference to PlayerConfig
	Input             *InputFrame           // Current frame inputs
	History           [][]byte              // List of serialized GameStates for replay
}

func (s *GameSession) Initialize() {
	// TODO: add this and needed functions for GameSession, GameState
}

func NewGameSession() *GameSession {
	data, err := os.ReadFile(filepath.Join(BaseSystemPath, AssetsDir, GPConfig.MapPath))
	if err != nil {
		LogError("Failed to load map file: %v", err)
		return nil
	}
	mapData := NewMapFromString(string(data))
	players := make(map[int]*PlayerSnake)
	for pname := range PConfigs {
		id := int(GetUniqueID())
		players[id] = &PlayerSnake{
			BaseSnake: &BaseSnake{
				Body:           &CollisionTiles{Tiles: []Vec2i{}},
				Facing:         DirRight,
				NextFacing:     DirRight,
				Fett:           GPConfig.StartSnakeLength - 1,
				ticksSinceMove: 0,
				StatusEffects:  []StatusEffect{},
			},
			ID:       id,
			Config:   PConfigs[pname],
			HeldItem: ItemNone,
		}
		LogInfo("Registered player %s with ID %d", pname, id)
	}
	state := &GameState{
		Tick:     0,
		Map:      mapData,
		Players:  players,
		Apples:   []*Apple{},
		Items:    []*Item{},
		Entities: []Entity{},
		Events:   []*GameEvent{},
	}
	pconfigs := make(map[int]*PlayerConfig)
	for id, player := range players {
		pconfigs[id] = player.Config
	}
	session := &GameSession{
		State:             state,
		RegisteredPlayers: pconfigs,
		Input:             &InputFrame{},
		History:           [][]byte{},
	}
	session.Initialize()
	return session
}
