package game

import (
	"fmt"
	"math"
)

type BotSnake struct {
	*BaseSnake
	*EntityBase
}

func (b *BotSnake) CalculateNextDirection(state *GameState) {
	// calculate next direction
	// find closest player snake (not Owner)
	// do A* pathfinding to the tile in front of the closest player snake's head
	distances := state.DistanceToPlayers(b.Body.Tiles[0])
	closestPlayerID := -1
	closestDistance := math.MaxInt64
	for id, dist := range distances {
		if id != b.OwnerID && state.Players[id].OwnLayers().CollidesWith(b.ScanLayers()) && dist < closestDistance {
			closestDistance = dist
			closestPlayerID = id
		}
	}
	if closestPlayerID != -1 {
		player := state.Players[closestPlayerID]
		if player != nil && player.Body != nil && len(player.Body.Tiles) > 0 {
			targetTile := player.Body.Tiles[0].Add(player.Facing)
			path, found := state.AstarPath(b.Body.Tiles[0], targetTile, b.ScanLayers())
			if found && len(path) > 1 {
				nextDir := path[1].Sub(b.Body.Tiles[0])
				b.Facing = nextDir
				b.NextFacing = nextDir
			}
		}
	}
}

func (b *BotSnake) Update(state *GameState, hist *HistoryData) {
	b.EntityBase.Update(state, hist)
	if b.IsExpired() {
		return
	}
	if b.BaseSnake.ticksSinceMove == 0 {
		b.CalculateNextDirection(state)
	}
	b.BaseSnake.Update(state, hist)
}
func (b *BotSnake) OwnLayers() CollisionMask     { return LayerSnake }
func (b *BotSnake) ScanLayers() CollisionMask    { return LayerWall | LayerSnake }
func (b *BotSnake) GetCollider() CollisionObject { return b.BaseSnake.GetCollider() }
func (b *BotSnake) GetOwner() interface{}        { return b }
func (b *BotSnake) CanSelfCollide() bool         { return false }

func (b *BotSnake) HandleOtherCollisions(other Collidable, tile Vec2i, state *GameState) {
	other_owner := other.GetOwner()

	switch o := other_owner.(type) {
	case *PlayerSnake:
		if o.ID != b.OwnerID || b.LifeTime < int(float64(GConfig.TPS)*(GPConfig.BotDuration-GPConfig.BotSpawningTime)) {
			if tile.Equals(b.Body.Tiles[0]) {
				b.MarkForDeath("collided with player snake")
			}
			if o.Body.Tiles[0].Equals(tile) {
				o.MarkForDeath(fmt.Sprintf("killed by bot of player %d", b.OwnerID))
			}
		}
	case *BotSnake:
		if tile.Equals(b.Body.Tiles[0]) {
			b.MarkForDeath("collided with another bot")
		}
	}
}
func (b *BotSnake) OnCollision(other Collidable, tile Vec2i, state *GameState) {
	if !b.CheckSelfCollision(other, state) {
		if !b.CheckWallCollision(other, state) {
			b.HandleOtherCollisions(other, tile, state)
		}
	}
}

func NewBotSnake(pos Vec2i, dir Vec2i, ownerID int) *BotSnake {
	b := &BotSnake{
		EntityBase: &EntityBase{
			ID:       GetUniqueID(),
			Type:     EntityBot,
			Collider: nil,
			OwnerID:  ownerID,
			LifeTime: int(float64(GConfig.TPS) * GPConfig.BotDuration),
		},
	}
	bs := &BaseSnake{
		Body:           &CollisionTiles{Tiles: []Vec2i{pos}},
		Facing:         dir,
		NextFacing:     dir,
		Fett:           GPConfig.BotLength - 1,
		ticksSinceMove: 0,
		StatusEffects:  []*StatusEffect{},
		die: func(s *BaseSnake, reason string, state *GameState, hist *HistoryData) {
			s.StatusEffects = []*StatusEffect{NewDeadStatusEffect()}
			b.LifeTime = 0
		},
		owner: b,
	}
	bs.BaseSpeed = GPConfig.BotSpeed
	b.BaseSnake = bs
	return b
}
