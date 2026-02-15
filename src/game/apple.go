package game

type Apple struct {
	*EntityBase
	Nutrition  int  `msgpack:"-"`
	IsConsumed bool `msgpack:"-"`
}

func (a *Apple) Update(state *GameState, hist *HistoryData) {}
func (a *Apple) OwnLayers() CollisionMask                   { return LayerApple }
func (a *Apple) GetOwner() interface{}                      { return a }

func NewApple(id uint64, pos Vec2i) *Apple {
	return &Apple{
		EntityBase: &EntityBase{
			ID:   id,
			Type: EntityApple,
			Collider: &CollisionTiles{
				Tiles: []Vec2i{pos},
			},
			OwnerID:  -1,
			LifeTime: -1,
		},
		Nutrition: GPConfig.AppleNutrition,
	}
}
