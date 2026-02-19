package game

import "fmt"

type FartEntity struct {
	*EntityBase
	Center               Vec2i
	ticksSinceLastDamage int
	collidingPlayers     []int
}

func (f *FartEntity) OnCollision(other Collidable, tile Vec2i, state *GameState) {
	if !f.IsExpired() {
		if pl, ok := other.GetOwner().(*PlayerSnake); ok {
			if pl.ID != f.OwnerID {
				col, _ := f.Collider.IsColliding(&CollisionTiles{Tiles: []Vec2i{pl.Body.Tiles[0]}})
				if col {
					f.collidingPlayers = append(f.collidingPlayers, pl.ID)
				}
			}
		}
	}
}
func (f *FartEntity) OwnLayers() CollisionMask  { return LayerNone }
func (f *FartEntity) ScanLayers() CollisionMask { return LayerSnake }
func (f *FartEntity) GetOwner() interface{}     { return f }

func (f *FartEntity) Update(state *GameState, hist *HistoryData) {
	f.EntityBase.Update(state, hist)
	if !f.IsExpired() {
		f.ticksSinceLastDamage++
		if f.ticksSinceLastDamage >= int(float64(GConfig.TPS)/GPConfig.FartDamagePerSecond) {
			f.ticksSinceLastDamage = 0
			for _, playerID := range f.collidingPlayers {
				if pl, ok := state.Players[playerID]; ok {
					pl.RemoveTiles(1, fmt.Sprintf("suffocated by player %d", f.OwnerID))
				}
			}
		}
		f.collidingPlayers = make([]int, 0)
	}
}

func NewFart(ownerID int, center Vec2i) *FartEntity {
	f := &FartEntity{
		EntityBase: &EntityBase{
			ID:   GetUniqueID(),
			Type: EntityFartCloud,
			Collider: NewCollisionRectangle(
				center.Sub(Vec2i{int16(GPConfig.FartSize), int16(GPConfig.FartSize)}).MakeP(),
				1+2*GPConfig.FartSize, 1+2*GPConfig.FartSize,
				true,
			),
			OwnerID:  ownerID,
			LifeTime: int(float64(GConfig.TPS) * GPConfig.FartDuration),
		},
		Center:               center,
		ticksSinceLastDamage: 0,
	}
	return f
}
