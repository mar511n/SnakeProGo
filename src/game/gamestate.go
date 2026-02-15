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
	for _, entity := range s.Entities {
		if b, ok := entity.(*BulletEntity); ok {
			bullets = append(bullets, b)
		}
	}
	ag := &struct {
		PlayersIDs []int
		Players    []*PlayerSnake
		Apples     []*Apple
		Items      []*Item
		Events     []*GameEvent
		Bullets    []*BulletEntity
	}{
		PlayersIDs: ids,
		Players:    pls,
		Apples:     s.Apples,
		Items:      s.Items,
		Events:     s.Events,
		Bullets:    bullets,
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
