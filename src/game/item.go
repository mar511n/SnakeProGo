package game

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
	ItemType   ItemType
	IsConsumed bool
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
				Tiles: []Vec2i{pos},
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
