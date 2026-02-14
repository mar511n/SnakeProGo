package game

import (
	"fmt"
)

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
	Die            func(s *BaseSnake, reason string, state *GameState)
	Owner          interface{}
	markedForDeath string
}

func NewBaseSnake(spawnpoint Vec2i, direction Vec2i, length int) *BaseSnake {
	bs := &BaseSnake{
		Body:           &CollisionTiles{Tiles: []Vec2i{spawnpoint}},
		Facing:         direction,
		NextFacing:     direction,
		Fett:           length - 1,
		ticksSinceMove: 0,
		StatusEffects:  []StatusEffect{},
		Die: func(s *BaseSnake, reason string, state *GameState) {
			s.StatusEffects = []StatusEffect{&DeadEffect{}}
		},
	}
	bs.Owner = bs
	return bs
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
		if s.markedForDeath != "" {
			s.Die(s, s.markedForDeath, state)
			s.markedForDeath = ""
			return
		}
		speed_multiplier := s.UpdateEffects(state)
		s.UpdateMovement(state, speed_multiplier)
	}
}
func (s *BaseSnake) IsDead() bool {
	return len(s.StatusEffects) == 1 && s.StatusEffects[0].GetType() == StatusEffectDead
}
func (s *BaseSnake) CheckSelfCollision(other Collidable, state *GameState) (consumed bool) {
	if other.GetCollider() == s.GetCollider() {
		head_tile := s.Body.Tiles[0]
		for _, body_tile := range s.Body.Tiles[1:] {
			if head_tile.Equals(body_tile) {
				s.markedForDeath = "self collision"
				return true
			}
		}
		return true
	}
	return false
}
func (s *BaseSnake) CheckWallCollision(other Collidable, state *GameState) (consumed bool) {
	if _, ok := other.GetCollider().(*CollisionMap); ok {
		s.markedForDeath = "wall collision"
		return true
	}
	return false
}
func (s *BaseSnake) HandleOtherCollisions(other Collidable, tile Vec2i, state *GameState) {
	other_owner := other.GetOwner()

	switch o := other_owner.(type) {
	case *BaseSnake:
		// handle snake-snake collision
		if tile.Equals(o.Body.Tiles[0]) {
			// own head collided with other snake (or both heads collided)
			s.markedForDeath = "snake collision"
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
	if !s.CheckSelfCollision(other, state) {
		if !s.CheckWallCollision(other, state) {
			s.HandleOtherCollisions(other, tile, state)
		}
	}
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
	ID         int
	Config     *PlayerConfig // Reference to existing PlayerConfig struct (name, keys, stats)
	HeldItem   ItemType      // Currently held item (ItemNone if empty)
	InputQueue []PlayerActionTurn
}

func NewPlayerSnake(base *BaseSnake, id int, config *PlayerConfig) *PlayerSnake {
	base.Die = DiePlayer
	sn := &PlayerSnake{
		BaseSnake:  base,
		ID:         id,
		Config:     config,
		HeldItem:   ItemNone,
		InputQueue: make([]PlayerActionTurn, 0, 20),
	}
	sn.Owner = sn
	return sn
}

func (s *PlayerSnake) GetOwner() interface{} { return s }

func (s *PlayerSnake) HandleInput(new_action PlayerActionTurn, state *GameState) {
	if len(s.InputQueue) < GPConfig.InputQueueSize && new_action != ActionNone {
		// append fresh action to input queue
		s.InputQueue = append(s.InputQueue, new_action)
	}

	if s.NextFacing != s.Facing {
		return
	}
	// take s.InputQueue[0], check if it's a valid direction change, and if so set NextFacing and pop it from the queue
	// if not valid, just pop it from the queue and check the next one until we find a valid one or the queue is empty
	next_action := ActionNone
	for len(s.InputQueue) > 0 {
		action := s.InputQueue[0]
		s.InputQueue = s.InputQueue[1:]
		if action.IsValid(s.Facing) {
			next_action = action
			break
		}
	}
	switch next_action {
	case ActionUp:
		s.NextFacing = Vec2i{X: 0, Y: -1}
	case ActionDown:
		s.NextFacing = Vec2i{X: 0, Y: 1}
	case ActionLeft:
		s.NextFacing = Vec2i{X: -1, Y: 0}
	case ActionRight:
		s.NextFacing = Vec2i{X: 1, Y: 0}
	case ActionTurnLeft:
		s.NextFacing = s.Facing.Rotate90(-1)
	case ActionTurnRight:
		s.NextFacing = s.Facing.Rotate90(1)
	default:
		s.NextFacing = s.Facing
	}
}

func (s *PlayerSnake) UseItem(state *GameState) {
	if s.HeldItem != ItemNone {
		if handler, ok := ItemRegistry[s.HeldItem]; ok {
			handler(s.ID, state)
			LogInfo("PlayerSnake %d used item %v", s.ID, s.HeldItem)
		} else {
			LogWarning("No handler found for item type %v", s.HeldItem)
		}
		s.HeldItem = ItemNone
	}
}

// override Die to log player ID and check for revive item
func DiePlayer(si *BaseSnake, reason string, state *GameState) {
	// TODO: add death to stats
	// TODO: handle ghost behavior
	PlaySound("Dead")
	s, ok := si.Owner.(*PlayerSnake)
	if !ok {
		LogWarning("DiePlayer called on BaseSnake with non-PlayerSnake owner")
	}
	LogInfo("PlayerSnake %d died: %s", s.ID, reason)
	s.StatusEffects = []StatusEffect{&DeadEffect{}}
	// Check for revive item before marking as dead
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
		if tile.Equals(s.Body.Tiles[0]) {
			// own head collided with other snake (or both heads collided)
			s.markedForDeath = fmt.Sprintf("snake collision with %d", o.ID)
		} else if tile.Equals(o.Body.Tiles[0]) {
			// other snake's head collided with own body - add a kill to own score
			LogInfo("Snake %v killed snake %v by collision at %v", s.ID, o.ID, tile)
		}
	case *Apple:
		PlaySound("Eating")
		s.Fett += o.Nutrition
		o.IsConsumed = true
	case *Item:
		PlaySound("Item")
		s.HeldItem = o.ItemType
		o.IsConsumed = true
	default:
		LogInfo("Unhandled collision at %v with object of type %v", tile, other_owner)
	}
}
func (s *PlayerSnake) OnCollision(other Collidable, tile Vec2i, state *GameState) {
	if !s.BaseSnake.CheckSelfCollision(other, state) {
		if !s.BaseSnake.CheckWallCollision(other, state) {
			s.HandleOtherCollisions(other, tile, state)
		}
	}
}
