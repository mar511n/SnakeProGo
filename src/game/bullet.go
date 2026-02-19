package game

type BulletEntity struct {
	*EntityBase
	Trail              []Vec2i
	Dir                Vec2i
	pos                Vec2i
	update_period      int
	ticks_since_update int
}

func (b *BulletEntity) GetOwner() interface{} { return b }

func (b *BulletEntity) Update(state *GameState, hist *HistoryData) {
	b.EntityBase.Update(state, hist)
	if !b.IsExpired() {
		b.ticks_since_update++
		if b.ticks_since_update >= b.update_period {
			b.ticks_since_update = 0
			b.pos = b.pos.Add(b.Dir).MakeP()
			b.Trail = append(b.Trail, b.pos)
			b.Collider.Tiles = []Vec2i{b.pos}
		}
	}
}

func NewBullet(ownerID int, pos, dir Vec2i) *BulletEntity {
	b := &BulletEntity{
		EntityBase: &EntityBase{
			ID:       GetUniqueID(),
			Type:     EntityBullet,
			Collider: &CollisionTiles{Tiles: []Vec2i{pos}},
			OwnerID:  ownerID,
			LifeTime: int(float64(GConfig.TPS*GPConfig.BulletRange) / GPConfig.BulletSpeed),
		},
		Trail:              []Vec2i{pos},
		Dir:                dir,
		pos:                pos,
		update_period:      int(float64(GConfig.TPS) / GPConfig.BulletSpeed),
		ticks_since_update: 0,
	}
	return b
}
