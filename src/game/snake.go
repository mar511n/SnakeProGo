package game

import "fmt"

// BaseSnake contains the physical properties and status of a snake entity.
// Implements Collidable and Updatable.
// Update method handles movement, growth, and status effects updates.
type BaseSnake struct {
	Body           *CollisionTiles // Head is at index 0 of the Vec2i slice
	Facing         Vec2i           // Current movement direction
	NextFacing     Vec2i           // Buffered input direction
	Fett           int             // "Fett" counter for growth buffer
	StatusEffects  []StatusEffect  // Active status effects (e.g. dead, invincible, speed boost)
	ticksSinceMove int             // Counter to track movement timing based on speed
}

func SpawnSnakeAt(point Vec2i, direction Vec2i, length int) *BaseSnake {
	return &BaseSnake{
		Body:           &CollisionTiles{Tiles: []Vec2i{point}},
		Facing:         direction,
		NextFacing:     direction,
		Fett:           length - 1,
		ticksSinceMove: 0,
		StatusEffects:  []StatusEffect{},
	}
}

func (s *BaseSnake) UpdateEffects(state *GameState) (speed_multiplier float64) {
	speed_multiplier = 1.0
	new_status_effects := []StatusEffect{}
	for _, effect := range s.StatusEffects {
		effect.Update(state)
		if !effect.IsExpired() {
			new_status_effects = append(new_status_effects, effect)
			if effect.GetType() == StatusEffectSpeedBoost {
				if sb, ok := effect.(*SpeedBoostEffect); ok {
					speed_multiplier *= sb.Multiplier
				}
			}
		}
	}
	s.StatusEffects = new_status_effects
	return speed_multiplier
}

func (s *BaseSnake) UpdateMovement(state *GameState, speed_multiplier float64) {
	current_speed := GPConfig.SnakeSpeed * speed_multiplier
	ticks_per_move := float64(GConfig.TPS) / current_speed
	s.ticksSinceMove++
	if s.ticksSinceMove >= int(ticks_per_move) {
		s.ticksSinceMove = 0

		// Update Facing based on NextFacing if it's a valid direction change
		if s.NextFacing.X != -s.Facing.X || s.NextFacing.Y != -s.Facing.Y {
			s.Facing = s.NextFacing
		}

		new_head := Vec2i{
			X: s.Body.Tiles[0].X + s.Facing.X,
			Y: s.Body.Tiles[0].Y + s.Facing.Y,
		}

		if s.Fett > 0 {
			s.Fett--
			s.Body.Tiles = append([]Vec2i{new_head}, s.Body.Tiles...)
		} else {
			s.Body.Tiles = append([]Vec2i{new_head}, s.Body.Tiles[:len(s.Body.Tiles)-1]...)
		}
	}
}
func (s *BaseSnake) Update(state *GameState) {
	if s.IsDead() {
		s.UpdateEffects(state)
	} else {
		speed_multiplier := s.UpdateEffects(state)
		s.UpdateMovement(state, speed_multiplier)
	}
}
func (s *BaseSnake) IsDead() bool {
	return len(s.StatusEffects) == 1 && s.StatusEffects[0].GetType() == StatusEffectDead
}
func (s *BaseSnake) Die(reason string, state *GameState) {
	LogInfo("Snake %v died: %s", s.GetOwner(), reason)
	// mark as dead by removing all other StatusEffects and adding a DeadEffect (which will handle possible respawning as a ghost)
	s.StatusEffects = []StatusEffect{&DeadEffect{}}
}
func (s *BaseSnake) CheckSelfCollision(other Collidable, state *GameState) {
	if other.GetCollider() == s.GetCollider() {
		head_tile := s.Body.Tiles[0]
		for _, body_tile := range s.Body.Tiles[1:] {
			if head_tile.Equals(body_tile) {
				s.Die("self collision", state)
				return
			}
		}
		return
	}
}
func (s *BaseSnake) CheckWallCollision(other Collidable, state *GameState) {
	if _, ok := other.GetCollider().(*CollisionMap); ok {
		s.Die("wall collision", state)
		return
	}
}
func (s *BaseSnake) HandleOtherCollisions(other Collidable, tile Vec2i, state *GameState) {
	other_owner := other.GetOwner()

	switch o := other_owner.(type) {
	case *BaseSnake:
		// handle snake-snake collision
		if tile.Equals(o.Body.Tiles[0]) {
			// own head collided with other snake (or both heads collided)
			s.Die("snake collision", state)
		} else {
			// other snake's head collided with own body - add a kill to own score
			LogInfo("Snake %v killed snake %v by collision at %v", s.GetOwner(), o.GetOwner(), tile)
		}
	case *Apple:
		s.Fett += o.Nutrition
		o.IsConsumed = true
	}
}
func (s *BaseSnake) OnCollision(other Collidable, tile Vec2i, state *GameState) {
	s.CheckSelfCollision(other, state)
	s.CheckWallCollision(other, state)
	s.HandleOtherCollisions(other, tile, state)
}
func (s *BaseSnake) OwnLayers() CollisionMask {
	if s.IsDead() {
		return LayerNone
	}
	return LayerSnake
}
func (s *BaseSnake) ScanLayers() CollisionMask {
	if s.IsDead() {
		return LayerNone
	}
	return LayerSnake | LayerApple | LayerWall | LayerEntity | LayerItem
}
func (s *BaseSnake) GetCollider() CollisionObject { return s.Body }
func (s *BaseSnake) GetOwner() interface{}        { return s }
func (s *BaseSnake) CanSelfCollide() bool         { return true }

// PlayerSnake represents a player-controlled snake. Implements InputHandler and embeds BaseSnake.
type PlayerSnake struct {
	*BaseSnake
	ID       int
	Config   *PlayerConfig // Reference to existing PlayerConfig struct (name, keys, stats)
	HeldItem ItemType      // Currently held item (ItemNone if empty)
}

func (s *PlayerSnake) HandleInput(action PlayerAction, state *GameState) {
	switch action {
	case ActionUp:
		s.NextFacing = Vec2i{X: 0, Y: -1}
	case ActionDown:
		s.NextFacing = Vec2i{X: 0, Y: 1}
	case ActionLeft:
		s.NextFacing = Vec2i{X: -1, Y: 0}
	case ActionRight:
		s.NextFacing = Vec2i{X: 1, Y: 0}
	case ActionTurnLeft:
		s.NextFacing = s.Facing.Rotate90(1)
	case ActionTurnRight:
		s.NextFacing = s.Facing.Rotate90(-1)
	}
}

func (s *PlayerSnake) UseItem(state *GameState) {
	if s.HeldItem != ItemNone {
		LogInfo("PlayerSnake %d used item %v", s.ID, s.HeldItem)
		ItemRegistry[s.HeldItem](s.ID, state)
		s.HeldItem = ItemNone
	}
}

// override Die to log player ID and check for revive item
func (s *PlayerSnake) Die(reason string, state *GameState) {
	// TODO: add death to stats
	// TODO: handle ghost behavior
	LogInfo("PlayerSnake %d died: %s", s.ID, reason)
	// Check for revive item before marking as dead
	s.BaseSnake.Die(reason, state)

	if s.HeldItem == ItemRevive {
		LogInfo("PlayerSnake %d used a revive item at death", s.ID)
		s.HeldItem = ItemNone
		ItemRegistry[ItemRevive](s.ID, state)
	}
}

func (s *PlayerSnake) HandleOtherCollisions(other Collidable, tile Vec2i, state *GameState) {
	other_owner := other.GetOwner()

	switch o := other_owner.(type) {
	case *PlayerSnake:
		// handle snake-snake collision
		if tile.Equals(o.Body.Tiles[0]) {
			// own head collided with other snake (or both heads collided)
			s.Die(fmt.Sprintf("snake collision with %d", o.ID), state)
		} else {
			// other snake's head collided with own body - add a kill to own score
			LogInfo("Snake %v killed snake %v by collision at %v", s.ID, o.ID, tile)
		}
	case *Apple:
		s.Fett += o.Nutrition
		o.IsConsumed = true
	case *Item:
		s.HeldItem = o.ItemType
		o.IsConsumed = true
	default:
		LogInfo("Unhandled collision at %v with object of type %v", tile, o)
	}
}
