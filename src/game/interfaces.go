package game

type StatusEffectType int

const (
	StatusEffectCustom StatusEffectType = iota
	StatusEffectInvincibility
	StatusEffectDead
	StatusEffectGhost
	StatusEffectSpeedBoost
)

type StatusEffect interface {
	Updatable
	GetType() StatusEffectType
	IsExpired() bool
}

type Updatable interface {
	Update(state *GameState)
}

type GameEvent struct {
	ID      uint64
	Type    string
	Payload interface{}
}

type Entity interface {
	Collidable
	Updatable
}

type Collidable interface {
	OnCollision(other Collidable, state *GameState) // Handle collision with another object
	OwnLayers() CollisionMask                       // The collision layers this object belongs to
	ScanLayers() CollisionMask                      // The collision layers this object scans for collisions against
	GetCollider() CollisionObject                   // Returns the geometric shape of the object
	GetOwner() interface{}                          // Returns a pointer to the object itself (for reference in collision handling)
}

type CollisionObject interface {
	IsColliding(other CollisionObject) bool
}
