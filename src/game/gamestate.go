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

func (s *GameState) CheckPointCollision(tile Vec2i, layer CollisionMask) bool {
	collObj := &CollisionTiles{Tiles: []Vec2i{tile}}
	if coll, _ := CheckCollision(layer, s.Map.OwnLayers(), collObj, s.Map.GetCollider()); coll {
		return true
	}
	for _, player := range s.Players {
		if coll, _ := CheckCollision(layer, player.OwnLayers(), collObj, player.GetCollider()); coll {
			return true
		}
	}
	for _, apple := range s.Apples {
		if apple != nil {
			if coll, _ := CheckCollision(layer, apple.OwnLayers(), collObj, apple.GetCollider()); coll {
				return true
			}
		}
	}
	for _, item := range s.Items {
		if item != nil {
			if coll, _ := CheckCollision(layer, item.OwnLayers(), collObj, item.GetCollider()); coll {
				return true
			}
		}
	}
	for _, entity := range s.Entities {
		if coll, _ := CheckCollision(layer, entity.OwnLayers(), collObj, entity.GetCollider()); coll {
			return true
		}
	}
	return false
}

func (s *GameState) SpawnItem() *Item {
	pos := RandomPosition(s.Map.Collider.Width, s.Map.Collider.Height).Add(s.Map.Collider.P0)
	for i := 0; i < 100; i++ {
		if !s.CheckPointCollision(pos, NewCollisionMaskAllLayers()) {
			break
		}
		pos = RandomPosition(s.Map.Collider.Width, s.Map.Collider.Height).Add(s.Map.Collider.P0)
	}
	return &Item{
		EntityBase: &EntityBase{
			ID:       GetUniqueID(),
			Type:     EntityItem,
			Collider: &CollisionTiles{Tiles: []Vec2i{pos}},
			OwnerID:  -1,
			LifeTime: -1,
		},
		ItemType:   GetRandomItemType(),
		IsConsumed: false,
	}
}

func (s *GameState) SpawnApple() *Apple {
	pos := RandomPosition(s.Map.Collider.Width, s.Map.Collider.Height).Add(s.Map.Collider.P0)
	for i := 0; i < 100; i++ {
		if !s.CheckPointCollision(pos, NewCollisionMaskAllLayers()) {
			break
		}
		pos = RandomPosition(s.Map.Collider.Width, s.Map.Collider.Height).Add(s.Map.Collider.P0)
	}
	return &Apple{
		EntityBase: &EntityBase{
			ID:       GetUniqueID(),
			Type:     EntityApple,
			Collider: &CollisionTiles{Tiles: []Vec2i{pos}},
			OwnerID:  -1,
			LifeTime: -1,
		},
		Nutrition:  GPConfig.AppleNutrition,
		IsConsumed: false,
	}
}

type GameSession struct {
	State             *GameState
	RegisteredPlayers map[int]*PlayerConfig // Key: ID, Value: Reference to PlayerConfig
	Input             *InputFrame           // Current frame inputs
	WindConditions    []CheckWinCondition   // Conditions to check for victory
	OnGameOver        func(winnerIDs []int) // Callback for game over, with winner IDs
	History           [][]byte              // List of serialized GameStates for replay
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
}
func (s *GameSession) Update() {
	s.State.Tick++
	// process input
	s.Input.Process(s.RegisteredPlayers)
	// update players
	for id, player := range s.State.Players {
		player.HandleInput(s.Input.Directions[id], s.State)
		player.Update(s.State)
	}
	// use items
	for id, player := range s.State.Players {
		if s.Input.ItemsUsed[id] && player.HeldItem != ItemNone {
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
	// check win conditions
	for _, condition := range s.WindConditions {
		game_over, winnerIDs := condition(s.State)
		if game_over {
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
		Input:             &InputFrame{},
		WindConditions:    []CheckWinCondition{CheckLastOneStanding},
		OnGameOver:        gameoverCallback,
		History:           [][]byte{},
	}
	session.Initialize()
	return session
}
