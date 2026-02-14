package game

import (
	"bytes"
	"encoding/gob"
)

func init() {
	gob.Register(&PlayerSnake{})
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
	Tick     uint64               `msgpack:"-"`
	Map      *MapData             `msgpack:"-"`
	Players  map[int]*PlayerSnake // Keyed by player ID
	Apples   []*Apple             // Collectible apples
	Items    []*Item              // Collectible items
	Entities []Entity             // Dynamic entities (Bullets, Farts, Bots)
	Events   []*GameEvent         // for audio and visual effects
}

// Might not work at later times, since not all interfaces are registered.
func (s *GameState) Encode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(s)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *GameState) Decode(data []byte) error {
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
