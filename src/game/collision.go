package game

import "fmt"

func ResolveCollision(a, b Collidable, state *GameState) bool {
	if a.ScanLayers().CollidesWith(b.OwnLayers()) {
		colliding, tile := a.GetCollider().IsColliding(b.GetCollider())
		if colliding {
			if a.GetCollider() == b.GetCollider() {
				if a.CanSelfCollide() {
					a.OnCollision(b, tile, state)
					return true
				}
			} else {
				a.OnCollision(b, tile, state)
				return true
			}
		}
	}
	return false
}

func CheckCollision(layerA, layerB CollisionMask, collObjA, collObjB CollisionObject) (bool, Vec2i) {
	if layerA.CollidesWith(layerB) {
		colliding, tile := collObjA.IsColliding(collObjB)
		if colliding {
			return true, tile
		}
	}
	return false, Vec2i{}
}

// CollisionTiles implements CollisionObject for sparse objects (Entities, Snakes)
type CollisionTiles struct {
	Tiles []Vec2i
}

func (c *CollisionTiles) IsColliding(other CollisionObject) (bool, Vec2i) {
	switch o := other.(type) {
	case *CollisionTiles:
		// Check for point overlaps O(N*M)
		for _, p1 := range c.Tiles {
			for _, p2 := range o.Tiles {
				if p1.Equals(p2) {
					return true, p1
				}
			}
		}
	case *CollisionMap:
		// Check if any point is in map bounds/walls
		for _, p := range c.Tiles {
			if o.Contains(p) {
				return true, p
			}
		}
	}
	return false, Vec2i{}
}

// CollisionMap implements CollisionObject for static map geometry
type CollisionMap struct {
	UseBounds     bool
	P0            Vec2i
	Width, Height int
	Occupied      [][]bool
}

func (c *CollisionMap) Contains(p Vec2i) bool {
	pr := p.Sub(c.P0)
	if pr.X < 0 || pr.Y < 0 || pr.X >= c.Width || pr.Y >= c.Height {
		return c.UseBounds
	}
	return c.Occupied[pr.X][pr.Y]
}

func (c *CollisionMap) IsColliding(other CollisionObject) (bool, Vec2i) {
	switch o := other.(type) {
	case *CollisionTiles:
		return o.IsColliding(c)
	}
	return false, Vec2i{}
}

type CollisionMask uint16

const (
	LayerNone   CollisionMask = 0
	LayerWall   CollisionMask = 1 << 0
	LayerSnake  CollisionMask = 1 << 1
	LayerApple  CollisionMask = 1 << 2
	LayerItem   CollisionMask = 1 << 3
	LayerEntity CollisionMask = 1 << 4
)

func (cm CollisionMask) AddLayer(layer CollisionMask) CollisionMask {
	return cm | layer
}

func (cm CollisionMask) RemoveLayer(layer CollisionMask) CollisionMask {
	return cm &^ layer
}

func NewCollisionMaskAllLayers() CollisionMask {
	return LayerWall | LayerSnake | LayerApple | LayerItem | LayerEntity
}

func (cm CollisionMask) CollidesWith(other CollisionMask) bool {
	if cm == LayerNone || other == LayerNone {
		return false
	}
	return cm&other != 0
}

func (cm CollisionMask) String() string {
	var layers []string
	if cm&LayerWall != 0 {
		layers = append(layers, "Wall")
	}
	if cm&LayerSnake != 0 {
		layers = append(layers, "Snake")
	}
	if cm&LayerApple != 0 {
		layers = append(layers, "Apple")
	}
	if cm&LayerItem != 0 {
		layers = append(layers, "Item")
	}
	if cm&LayerEntity != 0 {
		layers = append(layers, "Entity")
	}
	if len(layers) == 0 {
		return "None"
	}
	return fmt.Sprintf("%v", layers)
}
