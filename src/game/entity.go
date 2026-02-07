package game

type EntityType int

const (
	EntityApple EntityType = iota
	EntityItem
	EntityBullet
	EntityBomb
	EntityBot
	EntityFartCloud
)

// EntityBase represents any dynamic object in the world. Implements the Entity interface.
// Entities (Apples, Items, Bullets, Bots) embed this.
type EntityBase struct {
	ID       uint64
	Type     EntityType
	Collider *CollisionTiles // The entity's physical presence
	OwnerID  int             // Player who spawned it (or -1 for world)
	LifeTime int             // Ticks remaining (-1 for infinite)
}

func (e *EntityBase) Update(state *GameState) {
	if e.LifeTime > 0 {
		e.LifeTime--
	}
}
func (e *EntityBase) OnCollision(other Collidable, state *GameState) {}
func (e *EntityBase) OwnLayers() CollisionMask                       { return LayerEntity }
func (e *EntityBase) ScanLayers() CollisionMask                      { return LayerNone }
func (e *EntityBase) GetCollider() CollisionObject                   { return e.Collider }
func (e *EntityBase) GetOwner() interface{}                          { return e }

type Apple struct {
	*EntityBase
	Nutrition int
}

func (a *Apple) Update(state *GameState)  {}
func (a *Apple) OwnLayers() CollisionMask { return LayerApple }
func (a *Apple) GetOwner() interface{}    { return a }

func NewApple(id uint64, pos Vec2i) *Apple {
	return &Apple{
		EntityBase: &EntityBase{
			ID:   id,
			Type: EntityApple,
			Collider: &CollisionTiles{
				Points: []Vec2i{pos},
			},
			OwnerID:  -1,
			LifeTime: -1,
		},
		Nutrition: GPConfig.AppleNutrition,
	}
}

type ItemType int

const (
	ItemNone ItemType = iota
	ItemSpeed
	ItemRevive
	ItemShooting
	ItemBomb
	ItemBot
	ItemFart
)

// ItemHandler defines the effect of using an item.
// It returns true if the item was successfully used (and should be consumed).
type ItemHandler func(userID int, state *GameState) bool

// Global registry of item behaviors, populated at startup.
var ItemRegistry = map[ItemType]ItemHandler{}

type Item struct {
	*EntityBase
	ItemType ItemType
}

func (i *Item) Update(state *GameState)  {}
func (i *Item) OwnLayers() CollisionMask { return LayerItem }
func (i *Item) GetOwner() interface{}    { return i }

func NewItem(id uint64, pos Vec2i, itemType ItemType) *Item {
	return &Item{
		EntityBase: &EntityBase{
			ID:   id,
			Type: EntityItem,
			Collider: &CollisionTiles{
				Points: []Vec2i{pos},
			},
			OwnerID:  -1,
			LifeTime: -1,
		},
		ItemType: itemType,
	}
}

func init() {
	//TODO: This is where we would register all item behaviors.

	ItemRegistry[ItemSpeed] = func(userID int, state *GameState) bool {
		_, ok := state.Players[userID]
		if !ok {
			LogWarning("Player %d not found while trying to use Speed Item", userID)
			return false
		}
		LogInfo("Player %d used Speed Item", userID)
		state.Players[userID].StatusEffects = append(state.Players[userID].StatusEffects, &SpeedBoostEffect{
			Duration:   int(GPConfig.SpeedDuration * float64(GConfig.TPS)),
			Multiplier: GPConfig.SpeedMultiplier,
		})
		return true
	}
}
