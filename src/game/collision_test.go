package game

import (
	"testing"
)

func TestCollisionMask(t *testing.T) {
	cm := LayerNone
	t.Log(cm)
	cm = cm.AddLayer(LayerWall)
	t.Log(cm)
	cm = cm.AddLayer(LayerSnake)
	t.Log(cm)
	if !cm.CollidesWith(LayerWall) {
		t.Error("Expected collision with LayerWall")
	}
	if !cm.CollidesWith(LayerSnake) {
		t.Error("Expected collision with LayerSnake")
	}
	if cm.CollidesWith(LayerApple) {
		t.Error("Did not expect collision with LayerApple")
	}
	cm = cm.RemoveLayer(LayerWall)
	t.Log(cm)
	if cm.CollidesWith(LayerWall) {
		t.Error("Did not expect collision with LayerWall after removal")
	}
}

func TestCollisionTiles(t *testing.T) {
	tiles1 := &CollisionTiles{
		Points: []Vec2i{{X: 0, Y: 0}, {X: 1, Y: 1}},
	}
	tiles2 := &CollisionTiles{
		Points: []Vec2i{{X: 1, Y: 1}, {X: 2, Y: 2}},
	}
	tiles3 := &CollisionTiles{
		Points: []Vec2i{{X: 2, Y: 2}, {X: 3, Y: 3}},
	}

	if !tiles1.IsColliding(tiles2) {
		t.Error("Expected collision between tiles1 and tiles2")
	}

	if tiles1.IsColliding(tiles3) {
		t.Error("Did not expect collision between tiles1 and tiles3")
	}
}

func TestCollisionMap(t *testing.T) {
	// 5x5 map
	occupied := make([][]bool, 5)
	for i := range occupied {
		occupied[i] = make([]bool, 5)
	}
	occupied[2][2] = true // Obstacle at center

	cm := &CollisionMap{
		UseBounds: true,
		P0:        Vec2i{X: 0, Y: 0},
		Width:     5,
		Height:    5,
		Occupied:  occupied,
	}

	// Test Contains
	if !cm.Contains(Vec2i{X: -1, Y: 0}) {
		t.Error("Expected out of bounds point to be contained (collision)")
	}
	if !cm.Contains(Vec2i{X: 5, Y: 0}) {
		t.Error("Expected out of bounds point to be contained (collision)")
	}
	if !cm.Contains(Vec2i{X: 2, Y: 2}) {
		t.Error("Expected occupied point to be contained")
	}
	if cm.Contains(Vec2i{X: 0, Y: 0}) {
		t.Error("Expected empty point to not be contained")
	}

	// Test IsColliding with CollisionTiles
	tilesHit := &CollisionTiles{
		Points: []Vec2i{{X: 2, Y: 2}},
	}
	tilesMiss := &CollisionTiles{
		Points: []Vec2i{{X: 0, Y: 0}},
	}
	tilesBounds := &CollisionTiles{
		Points: []Vec2i{{X: -1, Y: -1}},
	}

	if !cm.IsColliding(tilesHit) {
		t.Error("Expected collision with occupied tile")
	}
	if cm.IsColliding(tilesMiss) {
		t.Error("Did not expect collision with empty tile")
	}
	if !cm.IsColliding(tilesBounds) {
		t.Error("Expected collision with out of bounds tile")
	}
}
