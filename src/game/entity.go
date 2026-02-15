package game

type EntityType int

const (
	EntityBasic EntityType = iota
	EntityApple
	EntityItem
	EntityBullet
	EntityBomb
	EntityBot
	EntityFartCloud
)

// EntityBase represents any dynamic object in the world. Implements the Entity interface.
// Entities (Apples, Items, Bullets, Bots) embed this.
type EntityBase struct {
	ID       uint64 `msgpack:"-"`
	Type     EntityType
	Collider *CollisionTiles // The entity's physical presence
	OwnerID  int             `msgpack:"-"` // Player who spawned it (or -1 for world)
	LifeTime int             `msgpack:"-"` // Ticks remaining (-1 for infinite)
}

func (e *EntityBase) Update(state *GameState, hist *HistoryData) {
	if e.LifeTime > 0 {
		e.LifeTime--
	}
}
func (e *EntityBase) OnCollision(other Collidable, tile Vec2i, state *GameState) {}
func (e *EntityBase) OwnLayers() CollisionMask                                   { return LayerEntity }
func (e *EntityBase) ScanLayers() CollisionMask                                  { return LayerNone }
func (e *EntityBase) GetCollider() CollisionObject                               { return e.Collider }
func (e *EntityBase) GetOwner() interface{}                                      { return e }
func (e *EntityBase) CanSelfCollide() bool                                       { return false }
