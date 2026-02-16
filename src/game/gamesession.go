package game

import (
	"os"
	"path/filepath"
	"time"
)

type GameSession struct {
	State             *GameState
	RegisteredPlayers map[int]*PlayerConfig // Key: ID, Value: Reference to PlayerConfig
	//Input             *InputFrame           // Current frame inputs
	WindConditions []CheckWinCondition                      // Conditions to check for victory
	OnGameOver     func(winnerIDs []int, hist *HistoryData) // Callback for game over, with winner IDs
	History        *HistoryData
	inputprocessor InputProcessor
	updatetimes    []time.Duration
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
	s.History.Init(s.State)
	s.updatetimes = make([]time.Duration, 0, int(GConfig.TPS)*10)
}
func (s *GameSession) Update() {
	s.State.Tick++

	startTime := time.Now()
	// process input
	input := s.inputprocessor(s.RegisteredPlayers)
	// update players
	for id, player := range s.State.Players {
		player.HandleInput(input.Directions[id], s.State)
		player.Update(s.State, s.History)
	}
	// use items
	for id, player := range s.State.Players {
		if input.ItemsUsed[id] && player.HeldItem != ItemNone {
			player.UseItem(s.State, s.History)
		}
	}
	// update entities
	newEntities := []Entity{}
	for _, entity := range s.State.Entities {
		entity.Update(s.State, s.History)
		if !entity.IsExpired() {
			newEntities = append(newEntities, entity)
		}
	}
	s.State.Entities = newEntities

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

	s.History.AddEntry(s.State)
	s.State.Events = []*GameEvent{}

	elapsed := time.Since(startTime)
	s.updatetimes = append(s.updatetimes, elapsed)
	if len(s.updatetimes) > int(GConfig.TPS)*10 {
		s.updatetimes = s.updatetimes[1:]
	}
	relElapsed := int(elapsed.Seconds() * float64(GConfig.TPS) * 100)
	if relElapsed > 80 {
		LogWarning("Update took above 80%% of tick duration! (%v)", elapsed)
	}

	if int(s.State.Tick)%(GConfig.TPS*10) == 0 {
		avgUpdateTime := time.Duration(0)
		maxUpdateTime := time.Duration(0)
		for _, t := range s.updatetimes {
			avgUpdateTime += t
			if t > maxUpdateTime {
				maxUpdateTime = t
			}
		}
		avgUpdateTime /= time.Duration(len(s.updatetimes))
		relavgUpdateTime := int(avgUpdateTime.Seconds() * float64(GConfig.TPS) * 100)
		relmaxUpdateTime := int(maxUpdateTime.Seconds() * float64(GConfig.TPS) * 100)
		LogInfo("Tick: %d, Avg Update Time: %v (%v/100), Max Update Time: %v (%v/100), Last Update Time: %v", s.State.Tick, avgUpdateTime, relavgUpdateTime, maxUpdateTime, relmaxUpdateTime, elapsed)
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
				s.OnGameOver(winnerIDs, s.History)
			}
			return
		}
	}
}

func NewGameSession(gameoverCallback func(winnerIDs []int, hist *HistoryData)) *GameSession {
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
		players[id].StatusEffects = append(players[id].StatusEffects, NewInvincibilityStatusEffect(GPConfig.StartInvincibilityDuration))
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
		History:           &HistoryData{},
		inputprocessor:    DefaultInputProcessor,
	}
	session.Initialize()
	return session
}
