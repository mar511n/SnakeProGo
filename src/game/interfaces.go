package game

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type GuiElement interface {
	Drawable
	Updatable
	SetContextID(id int)
	GetContextID() int
	SetNeighboringElement(dir GuiDirection, element GuiElement)
	GetNeighboringElement(dir GuiDirection) (GuiElement, bool)
	Bounds() *Rectf
	OnFocusGained()
	OnFocusLost()
	OnSelect()
	OnDeselect()
	OnInput(dirAction GuiAction, selectAction GuiAction)
}

type Context interface {
	Updatable
	Drawable
	StartContextSwitch(newContext Context, switchContext func(toContext Context)) error
}

type InputProcessor func(playerConfigs map[int]*PlayerConfig) *InputFrame

type Renderer interface {
	InitRender(useAntialias bool, state *GameState)
	Render(state *GameState, screen *ebiten.Image)
}

type InputHandler interface {
	HandleInput(input string, state *GameState)
}

type UpdatableGameObj interface {
	Update(state *GameState, hist *HistoryData)
}

type Updatable interface {
	Update() error
}

type Drawable interface {
	Draw(screen *ebiten.Image)
}

type Entity interface {
	Collidable
	UpdatableGameObj
	IsExpired() bool
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
