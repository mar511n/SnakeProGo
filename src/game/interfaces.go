package game

import "github.com/hajimehoshi/ebiten/v2"

type Renderer interface {
	InitRender(useAntialias bool, state *GameState)
	Render(state *GameState, screen *ebiten.Image)
}

type InputHandler interface {
	HandleInput(input string, state *GameState)
}

type StatusEffect interface {
	Updatable
	GetType() StatusEffectType
	IsExpired() bool
}

type Updatable interface {
	Update(state *GameState)
}

type Entity interface {
	Collidable
	Updatable
}

type Collidable interface {
	OnCollision(other Collidable, tile Vec2i, state *GameState) // Handle collision with another object
	OwnLayers() CollisionMask                                   // The collision layers this object belongs to
	ScanLayers() CollisionMask                                  // The collision layers this object scans for collisions against
	GetCollider() CollisionObject                               // Returns the geometric shape of the object
	GetOwner() interface{}                                      // Returns a pointer to the object itself (for reference in collision handling)
	CanSelfCollide() bool                                       // Whether this object can collide with itself (e.g. for snake body segments)
}

type CollisionObject interface {
	IsColliding(other CollisionObject) (bool, Vec2i)
}
