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

func (it ItemType) FileName() string {
	switch it {
	case ItemNone:
		return "none"
	case ItemSpeed:
		return "speed"
	case ItemRevive:
		return "revive"
	case ItemShooting:
		return "shot"
	case ItemBomb:
		return "bomb"
	case ItemBot:
		return "bot"
	case ItemFart:
		return "fart"
	default:
		return "unknown"
	}
}

func (it ItemType) String() string {
	switch it {
	case ItemNone:
		return "None"
	case ItemSpeed:
		return "Speed Boost"
	case ItemRevive:
		return "Revive"
	case ItemShooting:
		return "Shooting"
	case ItemBomb:
		return "Bomb"
	case ItemBot:
		return "Bot"
	case ItemFart:
		return "Fart"
	default:
		return "Unknown"
	}
}

// ItemHandler defines the effect of using an item.
// It returns true if the item was successfully used (and should be consumed).
type ItemHandler func(userID int, state *GameState, hist *HistoryData) bool

// Global registry of item behaviors, populated at startup.
var ItemRegistry = map[ItemType]ItemHandler{}

var ItemChances map[ItemType]float64

type Item struct {
	*EntityBase
	ItemType   ItemType
	IsConsumed bool `msgpack:"-"`
}

func (i *Item) Update(state *GameState, hist *HistoryData) {}
func (i *Item) OwnLayers() CollisionMask                   { return LayerItem }
func (i *Item) GetOwner() interface{}                      { return i }

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

func GetRandomItemType() ItemType {
	chance_sum := 0.0
	for _, chance := range ItemChances {
		chance_sum += chance
	}
	r := RandomSource.Float64() * chance_sum
	curr := 0.0
	for itemType, chance := range ItemChances {
		curr += chance
		if r < curr {
			return itemType
		}
	}
	return ItemNone
}

func InitializeItems() {
	ItemChances = make(map[ItemType]float64)
	ItemRegistry = make(map[ItemType]ItemHandler)
	ItemChances[ItemSpeed] = GPConfig.ItemSpeedChance
	ItemChances[ItemRevive] = GPConfig.ItemReviveChance
	ItemChances[ItemShooting] = GPConfig.ItemShootingChance
	ItemChances[ItemFart] = GPConfig.ItemFartChance
	//TODO: This is where we would register all item behaviors.

	ItemRegistry[ItemSpeed] = func(userID int, state *GameState, hist *HistoryData) bool {
		_, ok := state.Players[userID]
		if !ok {
			LogWarning("Player %d not found while trying to use Speed Item", userID)
			return false
		}
		state.PlaySoundEffect("Speed")
		state.Players[userID].StatusEffects = append(state.Players[userID].StatusEffects, NewSpeedBoostStatusEffect(GPConfig.SpeedDuration))
		return true
	}
	ItemRegistry[ItemRevive] = func(userID int, state *GameState, hist *HistoryData) (consumed bool) {
		consumed = false
		player, ok := state.Players[userID]
		if !ok {
			LogWarning("Player %d not found while trying to use Revive Item", userID)
			return
		}
		state.PlaySoundEffect("Revive")
		if GPConfig.ReviveIsRewind {
			tick := state.Tick
			hist.ReconstructState(int(state.Tick)-int(GPConfig.RewindTime*float64(GConfig.TPS)), state)
			state.Tick = tick
			state.Events = append(state.Events, NewSoundEvent("Revive"))
			consumed = true
		} else {
			pos := RandomPosition(int(state.Map.Collider.Width), int(state.Map.Collider.Height)).Add(state.Map.Collider.P0)
			for i := 0; i < 100; i++ {
				if !state.CheckPointCollision(pos, NewCollisionMaskAllLayers()) {
					break
				}
				pos = RandomPosition(int(state.Map.Collider.Width), int(state.Map.Collider.Height)).Add(state.Map.Collider.P0)
			}
			player.Body.Tiles = []Vec2i{pos}
			player.Fett = GPConfig.SnakeSurvivalLength - 1
			player.StatusEffects = []*StatusEffect{NewRespawningStatusEffect(GPConfig.ReviveDuration)}
			consumed = true
		}
		return
	}
	ItemRegistry[ItemShooting] = func(userID int, state *GameState, hist *HistoryData) (consumed bool) {
		consumed = false
		pl, ok := state.Players[userID]
		if !ok {
			LogWarning("Player %d not found while trying to use Shooting Item", userID)
			return
		}
		state.PlaySoundEffect("Shooting")
		state.Entities = append(state.Entities, NewBullet(
			pl.ID,
			pl.Body.Tiles[0].Add(pl.Facing).MakeP(),
			pl.Facing,
		))
		consumed = true
		return
	}
	ItemRegistry[ItemFart] = func(userID int, state *GameState, hist *HistoryData) (consumed bool) {
		consumed = false
		pl, ok := state.Players[userID]
		if !ok {
			LogWarning("Player %d not found while trying to use Fart Item", userID)
			return
		}
		state.PlaySoundEffect("Farting")
		state.Entities = append(state.Entities, NewFart(
			pl.ID,
			pl.Body.Tiles[len(pl.Body.Tiles)-1],
		))
		consumed = true
		return
	}
}
