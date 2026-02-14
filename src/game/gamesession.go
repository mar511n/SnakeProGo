package game

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/vmihailenco/msgpack/v5"
)

type GameSession struct {
	State             *GameState
	RegisteredPlayers map[int]*PlayerConfig // Key: ID, Value: Reference to PlayerConfig
	//Input             *InputFrame           // Current frame inputs
	WindConditions []CheckWinCondition   // Conditions to check for victory
	OnGameOver     func(winnerIDs []int) // Callback for game over, with winner IDs
	History        [][]byte              // List of serialized GameStates for replay
	HistoryTicks   []uint64              // Corresponding ticks for each entry in History
	HistorySize    int                   // Number of bytes currently stored in History (for memory management)
	Encoder        *msgpack.Encoder
	inputprocessor InputProcessor
}

func (s *GameSession) Initialize() {
	InitializeItems()
	s.State.Apples = make([]*Apple, GPConfig.AppleCount)
	s.State.Items = make([]*Item, GPConfig.ItemCount)
	for i := 0; i < GPConfig.AppleCount; i++ {
		s.State.Apples[i] = s.State.SpawnApple()
	}
	for i := 0; i < GPConfig.ItemCount; i++ {
		s.State.Items[i] = s.State.SpawnItem()
	}
	s.Encoder = msgpack.GetEncoder()
	s.Encoder.SetOmitEmpty(true)
	s.Encoder.UseCompactInts(true)
	s.Encoder.UseCompactFloats(true)
	s.Encoder.UseArrayEncodedStructs(true)
}
func (s *GameSession) Update() {
	s.State.Tick++
	// process input
	input := s.inputprocessor(s.RegisteredPlayers)
	// update players
	for id, player := range s.State.Players {
		player.HandleInput(input.Directions[id], s.State)
		player.Update(s.State)
	}
	// use items
	for id, player := range s.State.Players {
		if input.ItemsUsed[id] && player.HeldItem != ItemNone {
			player.UseItem(s.State)
		}
	}
	// update entities
	for _, entity := range s.State.Entities {
		entity.Update(s.State)
	}
	// check collisions
	collidables := make([]Collidable, 1+len(s.State.Players)+len(s.State.Apples)+len(s.State.Items)+len(s.State.Entities))
	collidables[0] = s.State.Map
	idx := 1
	for _, player := range s.State.Players {
		collidables[idx] = player
		idx++
	}
	for _, apple := range s.State.Apples {
		collidables[idx] = apple
		idx++
	}
	for _, item := range s.State.Items {
		collidables[idx] = item
		idx++
	}
	for _, entity := range s.State.Entities {
		collidables[idx] = entity
		idx++
	}
	total_collisions := 0
	for i := 0; i < len(collidables); i++ {
		for j := 0; j < len(collidables); j++ {
			if ResolveCollision(collidables[i], collidables[j], s.State) {
				total_collisions++
			}
		}
	}

	// spawn new apples/items if needed
	for ai := range s.State.Apples {
		if s.State.Apples[ai].IsConsumed {
			s.State.Apples[ai] = s.State.SpawnApple()
		}
	}
	for ii := range s.State.Items {
		if s.State.Items[ii].IsConsumed {
			s.State.Items[ii] = s.State.SpawnItem()
		}
	}

	//b, err := msgpack.Marshal(s.State)
	var buf bytes.Buffer
	s.Encoder.Reset(&buf)
	err := s.Encoder.Encode(s.State)
	b := buf.Bytes()
	msgpack.PutEncoder(s.Encoder)

	if err != nil {
		LogError("Failed to serialize game state for history: %v", err)
	} else {
		// check if b is different from the last entry in s.History to avoid storing duplicate states
		if len(s.History) == 0 || !bytes.Equal(b, s.History[len(s.History)-1]) {
			//LogInfo("Storing game state for history (tick %d, size %d)", s.State.Tick, len(b))
			s.History = append(s.History, b)
			s.HistoryTicks = append(s.HistoryTicks, s.State.Tick)
			s.HistorySize += len(b)
			// if history size exceeds max, remove half of the oldest entries
			if s.HistorySize > GConfig.MaxHistorySize {
				removeCount := len(s.History) / 2
				for i := 0; i < removeCount; i++ {
					s.HistorySize -= len(s.History[i])
				}
				s.History = s.History[removeCount:]
				s.HistoryTicks = s.HistoryTicks[removeCount:]
				LogInfo("History size exceeded max, removed %d oldest entries, new size %d", removeCount, s.HistorySize)
			}
		}
	}

	// check win conditions
	for _, condition := range s.WindConditions {
		game_over, winnerIDs := condition(s.State)
		if game_over {
			if len(winnerIDs) > 0 {
				s.State.PlaySoundEffect("Win")
			}
			LogInfo("Game over! Winners: %v", winnerIDs)
			if s.OnGameOver != nil {
				s.OnGameOver(winnerIDs)
			}
			return
		}
	}
}

func NewGameSession(gameoverCallback func(winnerIDs []int)) *GameSession {
	data, err := os.ReadFile(filepath.Join(BaseSystemPath, ResDir, GPConfig.MapPath))
	if err != nil {
		LogError("Failed to load map file: %v", err)
		return nil
	}
	mapData := NewMapFromString(string(data))
	players := make(map[int]*PlayerSnake)
	idx := 0
	for pname := range PConfigs {
		id := int(GetUniqueID())
		if idx >= len(mapData.SpawnPoints) {
			LogError("Not enough spawn points for players! Player %s cannot be spawned.", pname)
			continue
		}
		players[id] = NewPlayerSnake(
			NewBaseSnake(mapData.SpawnPoints[idx], mapData.SpawnDirs[idx], GPConfig.StartSnakeLength),
			id,
			PConfigs[pname],
		)
		idx++
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
		WindConditions:    []CheckWinCondition{CheckLastOneStanding},
		OnGameOver:        gameoverCallback,
		History:           [][]byte{},
		inputprocessor:    DefaultInputProcessor,
	}
	session.Initialize()
	return session
}
