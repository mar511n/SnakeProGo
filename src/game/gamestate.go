package game

import (
	"bytes"
	"encoding/gob"
	"sort"

	"github.com/vmihailenco/msgpack/v5"
)

func init() {
	gob.Register(&PlayerSnake{})
	gob.Register(&EntityBase{})
}

type GameEventType int

const (
	GameEventSound GameEventType = iota
	GameEventVisual
)

type GameEvent struct {
	ID      uint64 `msgpack:"-"`
	Type    GameEventType
	Payload interface{}
}

func NewSoundEvent(soundName string) *GameEvent {
	return &GameEvent{
		ID:      GetUniqueID(),
		Type:    GameEventSound,
		Payload: soundName,
	}
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

func (s *GameState) MarshalMsgpack() ([]byte, error) {
	ids := make([]int, 0, len(s.Players))
	pls := make([]*PlayerSnake, 0, len(s.Players))
	for id, _ := range s.Players {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		pls = append(pls, s.Players[id])
	}
	bullets := make([]*BulletEntity, 0)
	farts := make([]*FartEntity, 0)
	bots := make([]*BotSnake, 0)
	for _, entity := range s.Entities {
		if b, ok := entity.(*BulletEntity); ok {
			bullets = append(bullets, b)
		}
		if f, ok := entity.(*FartEntity); ok {
			farts = append(farts, f)
		}
		if b, ok := entity.(*BotSnake); ok {
			bots = append(bots, b)
		}
	}
	ag := &struct {
		PlayersIDs []int
		Players    []*PlayerSnake
		Apples     []*Apple
		Items      []*Item
		Events     []*GameEvent
		Bullets    []*BulletEntity
		Farts      []*FartEntity
		Bots       []*BotSnake
	}{
		PlayersIDs: ids,
		Players:    pls,
		Apples:     s.Apples,
		Items:      s.Items,
		Events:     s.Events,
		Bullets:    bullets,
		Farts:      farts,
		Bots:       bots,
	}
	b, err := msgpack.Marshal(ag)
	if err != nil {
		LogError("Failed to marshal GameState: %v", err)
	}
	return b, err
}

func (s *GameState) UnmarshalMsgpack(data []byte) error {
	aux := &struct {
		PlayersIDs []int
		Players    []*PlayerSnake
		Apples     []*Apple
		Items      []*Item
		Events     []*GameEvent
		Bullets    []*BulletEntity
		Farts      []*FartEntity
		Bots       []*BotSnake
	}{}
	err := msgpack.Unmarshal(data, aux)
	if err != nil {
		return err
	}
	if len(aux.PlayersIDs) != len(s.Players) {
		s.Players = make(map[int]*PlayerSnake)
	}
	for i, id := range aux.PlayersIDs {
		if _, exists := s.Players[id]; !exists {
			s.Players[id] = aux.Players[i]
		} else {
			s.Players[id].OverWriteWith(aux.Players[i])
		}
	}
	if aux.Apples != nil {
		s.Apples = aux.Apples
	}
	if aux.Items != nil {
		s.Items = aux.Items
	}
	if aux.Events != nil {
		s.Events = aux.Events
	}
	newEntities := make([]Entity, 0)
	if aux.Bullets != nil {
		for _, b := range aux.Bullets {
			newEntities = append(newEntities, b)
		}
	} else {
		for _, entity := range s.Entities {
			if _, ok := entity.(*BulletEntity); !ok {
				newEntities = append(newEntities, entity)
			}
		}
	}
	if aux.Farts != nil {
		for _, f := range aux.Farts {
			newEntities = append(newEntities, f)
		}
	} else {
		for _, entity := range s.Entities {
			if _, ok := entity.(*FartEntity); !ok {
				newEntities = append(newEntities, entity)
			}
		}
	}
	if aux.Bots != nil {
		for _, b := range aux.Bots {
			newEntities = append(newEntities, b)
		}
	} else {
		for _, entity := range s.Entities {
			if _, ok := entity.(*BotSnake); !ok {
				newEntities = append(newEntities, entity)
			}
		}
	}

	s.Entities = newEntities
	return nil
}

func (s *GameState) MarshalMutableObjects() ([]byte, error) {
	return msgpack.Marshal(s)
}

func (s *GameState) UnmarshalMutableObjects(data []byte) error {
	return msgpack.Unmarshal(data, s)
}

// Might not work at later times, since not all interfaces are registered.
func (s *GameState) MarshalAllObjects() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(s)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *GameState) UnmarshalAllObjects(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(s)
}

func (s *GameState) PlaySoundEffect(soundName string) {
	s.Events = append(s.Events, NewSoundEvent(soundName))
	PlaySound(soundName, GConfig.SfxVolume)
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
	pos := RandomPosition(int(s.Map.Collider.Width), int(s.Map.Collider.Height)).Add(s.Map.Collider.P0)
	for i := 0; i < 100; i++ {
		if !s.CheckPointCollision(pos, NewCollisionMaskAllLayers()) {
			break
		}
		pos = RandomPosition(int(s.Map.Collider.Width), int(s.Map.Collider.Height)).Add(s.Map.Collider.P0)
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
	pos := RandomPosition(int(s.Map.Collider.Width), int(s.Map.Collider.Height)).Add(s.Map.Collider.P0)
	for i := 0; i < 100; i++ {
		if !s.CheckPointCollision(pos, NewCollisionMaskAllLayers()) {
			break
		}
		pos = RandomPosition(int(s.Map.Collider.Width), int(s.Map.Collider.Height)).Add(s.Map.Collider.P0)
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

func (s *GameState) DistanceToPlayers(pos Vec2i) map[int]int {
	distances := make(map[int]int)
	for id, player := range s.Players {
		if player != nil && player.Body != nil && len(player.Body.Tiles) > 0 {
			headPos := player.Body.Tiles[0]
			//diff := pos.Sub(headPos)
			distances[id] = heuristic(pos, headPos) //math.Hypot(float64(diff.X), float64(diff.Y))
		}
	}
	return distances
}

// calculate the shortest path from start to goal using A* algorithm. Returns the path containing the start and goal tiles
// use s.CheckPointCollision(tile, scanLayers) to check if a tile is blocked
func (s *GameState) AstarPath(start, goal Vec2i, scanLayers CollisionMask) ([]Vec2i, bool) {
	// If start equals goal, return just the start
	if start.Equals(goal) {
		return []Vec2i{start}, true
	}

	// Check if start or goal are blocked
	if s.CheckPointCollision(goal, scanLayers) {
		return nil, false
	}

	// A* algorithm structures
	type node struct {
		pos    Vec2i
		g      int
		h      int
		f      int
		parent *node
	}

	// Directions: 4-directional movement (up, right, down, left)
	directions := []Vec2i{
		{0, -1}, // up
		{1, 0},  // right
		{0, 1},  // down
		{-1, 0}, // left
	}

	// Open set using a simple map and priority queue approach
	openSet := make(map[Vec2i]*node)
	closedSet := make(map[Vec2i]bool)

	// Start node
	startNode := &node{
		pos:    start,
		g:      0,
		h:      heuristic(start, goal),
		parent: nil,
	}
	startNode.f = startNode.g + startNode.h
	openSet[start] = startNode

	for len(openSet) > 0 {
		// Find node with lowest f score in open set
		var current *node
		minF := int(^uint(0) >> 1) // max int
		for _, n := range openSet {
			if n.f < minF {
				minF = n.f
				current = n
			}
		}

		// Check if we reached the goal
		if current.pos.Equals(goal) {
			// Reconstruct path
			path := []Vec2i{}
			for n := current; n != nil; n = n.parent {
				path = append([]Vec2i{n.pos}, path...)
			}
			return path, true
		}

		// Move current from open to closed set
		delete(openSet, current.pos)
		closedSet[current.pos] = true

		// Explore neighbors
		for _, dir := range directions {
			neighborPos := current.pos.Add(dir)

			// Skip if in closed set
			if closedSet[neighborPos] {
				continue
			}

			// Skip if blocked
			if s.CheckPointCollision(neighborPos, scanLayers) {
				continue
			}

			// Calculate g score for this neighbor
			tentativeG := current.g + 1 // uniform cost for each step

			// Check if neighbor is in open set
			neighborNode, exists := openSet[neighborPos]
			if !exists {
				// Create new node
				neighborNode = &node{
					pos:    neighborPos,
					g:      tentativeG,
					h:      heuristic(neighborPos, goal),
					parent: current,
				}
				neighborNode.f = neighborNode.g + neighborNode.h
				openSet[neighborPos] = neighborNode
			} else if tentativeG < neighborNode.g {
				// Found a better path to this node
				neighborNode.g = tentativeG
				neighborNode.f = neighborNode.g + neighborNode.h
				neighborNode.parent = current
			}
		}
	}

	// No path found
	return nil, false
}

// Heuristic function using Manhattan distance
func heuristic(a, b Vec2i) int {
	dx := int(a.X - b.X)
	dy := int(a.Y - b.Y)
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}
