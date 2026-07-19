package game

import (
	"fmt"
)

// BaseSnake contains the physical properties and status of a snake entity.
// Implements Collidable and UpdatableGameObj.
// Update method handles movement, growth, and status effects updates.
type BaseSnake struct {
	Body           *CollisionTiles // Head is at index 0 of the Vec2i slice
	Facing         Vec2i           // Current movement direction
	NextFacing     Vec2i           `msgpack:"-"` // Buffered input direction
	Fett           int             `msgpack:"-"` // "Fett" counter for growth buffer
	StatusEffects  []*StatusEffect // Active status effects (e.g. dead, invincible, speed boost)
	BaseSpeed      float64         `msgpack:"-"` // Base speed in segments per second
	ticksSinceMove int             // Counter to track movement timing based on speed
	die            func(s *BaseSnake, reason string, state *GameState, hist *HistoryData)
	owner          interface{}
	markedForDeath string
}

func (s *BaseSnake) MarkForDeath(reason string) bool {
	if s.markedForDeath == "" {
		s.markedForDeath = reason
		return true
	}
	return false
}

func (s *BaseSnake) OverWriteWith(other *BaseSnake) {
	if other.Body != nil {
		s.Body = other.Body
	}
	s.Facing = other.Facing
	s.NextFacing = other.Facing
	s.Fett = other.Fett
	if other.StatusEffects != nil {
		s.StatusEffects = other.StatusEffects
	}
}

func NewBaseSnake(spawnpoint Vec2i, direction Vec2i, length int) *BaseSnake {
	bs := &BaseSnake{
		Body:           &CollisionTiles{Tiles: []Vec2i{spawnpoint}},
		Facing:         direction,
		NextFacing:     direction,
		Fett:           length - 1,
		BaseSpeed:      GPConfig.SnakeSpeed,
		ticksSinceMove: 0,
		StatusEffects:  []*StatusEffect{},
		die: func(s *BaseSnake, reason string, state *GameState, hist *HistoryData) {
			s.StatusEffects = []*StatusEffect{NewDeadStatusEffect()}
		},
	}
	bs.owner = bs
	return bs
}

func (s *BaseSnake) UpdateEffects(state *GameState, hist *HistoryData) (speed_multiplier float64) {
	speed_multiplier = 1.0
	new_status_effects := []*StatusEffect{}
	for _, effect := range s.StatusEffects {
		effect.Update(s, state, hist)
		if !effect.IsExpired() {
			new_status_effects = append(new_status_effects, effect)
			if effect.Type == StatusEffectSpeedBoost {
				speed_multiplier *= GPConfig.SpeedMultiplier
			}
		}
	}
	s.StatusEffects = new_status_effects
	return speed_multiplier
}

func (s *BaseSnake) UpdateMovement(state *GameState, speed_multiplier float64) {
	current_speed := s.BaseSpeed * speed_multiplier
	ticks_per_move := float64(GConfig.TPS) / current_speed
	s.ticksSinceMove++
	if s.ticksSinceMove >= int(ticks_per_move) {
		s.ticksSinceMove = 0

		// Update Facing based on NextFacing if it's a valid direction change
		if s.NextFacing.X != -s.Facing.X || s.NextFacing.Y != -s.Facing.Y {
			s.Facing = s.NextFacing
		}

		new_head := s.Body.Tiles[0].Add(s.Facing).MakeP()

		if s.Fett > 0 {
			s.Fett--
			s.Body.Tiles = append([]Vec2i{new_head}, s.Body.Tiles...)
		} else {
			s.Body.Tiles = append([]Vec2i{new_head}, s.Body.Tiles[:len(s.Body.Tiles)-1]...)
		}
	}
}
func (s *BaseSnake) Update(state *GameState, hist *HistoryData) {
	if s.IsDead() || s.HasStatusEffect(StatusEffectRespawning) {
		s.UpdateEffects(state, hist)
	} else {
		if s.markedForDeath != "" {
			s.die(s, s.markedForDeath, state, hist)
			s.markedForDeath = ""
			return
		}
		speed_multiplier := s.UpdateEffects(state, hist)
		s.UpdateMovement(state, speed_multiplier)
	}
}
func (s *BaseSnake) IsDead() bool {
	return len(s.StatusEffects) == 1 && s.StatusEffects[0].Type == StatusEffectDead
}
func (s *BaseSnake) HasStatusEffect(effectType StatusEffectType) bool {
	for _, effect := range s.StatusEffects {
		if effect.Type == effectType {
			return true
		}
	}
	return false
}
func (s *BaseSnake) RemoveTiles(num int, death_reason string) {
	if len(s.Body.Tiles)-num < GPConfig.SnakeSurvivalLength {
		if !s.HasStatusEffect(StatusEffectInvincibility) && !s.IsDead() && !s.HasStatusEffect(StatusEffectRespawning) && !s.HasStatusEffect(StatusEffectGhost) {
			s.MarkForDeath(death_reason)
		}
	} else {
		if !s.HasStatusEffect(StatusEffectInvincibility) {
			s.Body.Tiles = s.Body.Tiles[:len(s.Body.Tiles)-num]
		}
	}
}
func (s *BaseSnake) CheckSelfCollision(other Collidable, state *GameState) (consumed bool) {
	if other.GetCollider() == s.GetCollider() {
		head_tile := s.Body.Tiles[0]
		for _, body_tile := range s.Body.Tiles[1:] {
			if head_tile.Equals(body_tile) {
				s.MarkForDeath("self collision")
				return true
			}
		}
		return true
	}
	return false
}
func (s *BaseSnake) CheckWallCollision(other Collidable, state *GameState) (consumed bool) {
	if _, ok := other.GetCollider().(*CollisionMap); ok {
		s.MarkForDeath("wall collision")
		return true
	}
	return false
}
func (s *BaseSnake) HandleOtherCollisions(other Collidable, tile Vec2i, state *GameState) {}
func (s *BaseSnake) OnCollision(other Collidable, tile Vec2i, state *GameState) {
	if !s.CheckSelfCollision(other, state) {
		if !s.CheckWallCollision(other, state) {
			s.HandleOtherCollisions(other, tile, state)
		}
	}
}
func (s *BaseSnake) OwnLayers() CollisionMask {
	if s.IsDead() || s.HasStatusEffect(StatusEffectRespawning) {
		return LayerNone
	}
	return LayerSnake
}
func (s *BaseSnake) ScanLayers() CollisionMask {
	if s.IsDead() || s.HasStatusEffect(StatusEffectRespawning) {
		return LayerNone
	} else if s.HasStatusEffect(StatusEffectInvincibility) {
		return LayerApple | LayerItem
	}
	return LayerWall | LayerSnake | LayerEntity | LayerApple | LayerItem
}
func (s *BaseSnake) GetCollider() CollisionObject { return s.Body }
func (s *BaseSnake) GetOwner() interface{}        { return s.owner }
func (s *BaseSnake) CanSelfCollide() bool         { return true }

// PlayerSnake represents a player-controlled snake. Implements InputHandler and embeds BaseSnake.
type PlayerSnake struct {
	*BaseSnake
	ID         int
	Config     *PlayerConfig      `msgpack:"-"` // Reference to existing PlayerConfig struct (name, keys, stats)
	HeldItem   ItemType           // Currently held item (ItemNone if empty)
	InputQueue []PlayerActionTurn `msgpack:"-"`
}

func (s *PlayerSnake) OverWriteWith(other *PlayerSnake) {
	s.BaseSnake.OverWriteWith(other.BaseSnake)
	s.ID = other.ID
	s.HeldItem = other.HeldItem
}

func NewPlayerSnake(base *BaseSnake, id int, config *PlayerConfig) *PlayerSnake {
	base.die = DiePlayer
	sn := &PlayerSnake{
		BaseSnake:  base,
		ID:         id,
		Config:     config,
		HeldItem:   ItemNone,
		InputQueue: make([]PlayerActionTurn, 0, 20),
	}
	sn.owner = sn
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

func (s *PlayerSnake) UseItem(state *GameState, hist *HistoryData) {
	if s.HeldItem != ItemNone {
		consumed := false
		if handler, ok := ItemRegistry[s.HeldItem]; ok {
			consumed = handler(s.ID, state, hist)
			LogInfo("PlayerSnake %d used item %v", s.ID, s.HeldItem)
		} else {
			LogWarning("No handler found for item type %v", s.HeldItem)
		}
		if consumed {
			s.HeldItem = ItemNone
		}
	}
}

// override Die to log player ID and check for revive item
func DiePlayer(si *BaseSnake, reason string, state *GameState, hist *HistoryData) {
	// TODO: add death to stats
	// TODO: handle ghost behavior
	s, ok := si.owner.(*PlayerSnake)
	if !ok {
		LogWarning("DiePlayer called on BaseSnake with non-PlayerSnake owner")
	}
	LogInfo("PlayerSnake %d died: %s", s.ID, reason)
	s.StatusEffects = []*StatusEffect{NewDeadStatusEffect()}
	// Check for revive item before marking as dead
	consumedRevive := false
	if s.HeldItem == ItemRevive {
		LogInfo("PlayerSnake %d used a revive item at death", s.ID)
		consumedRevive = ItemRegistry[ItemRevive](s.ID, state, hist)
	}
	if consumedRevive {
		s.HeldItem = ItemNone
	} else {
		state.PlaySoundEffect("Dead")
	}
}

func (s *PlayerSnake) HandleOtherCollisions(other Collidable, tile Vec2i, state *GameState) {
	other_owner := other.GetOwner()

	switch o := other_owner.(type) {
	case *PlayerSnake:
		// handle snake-snake collision
		if tile.Equals(s.Body.Tiles[0]) {
			// own head collided with other snake (or both heads collided)
			s.MarkForDeath(fmt.Sprintf("snake collision with %d", o.ID))
		} else if tile.Equals(o.Body.Tiles[0]) {
			// other snake's head collided with own body - add a kill to own score
			LogInfo("Snake %v killed snake %v by collision at %v", s.ID, o.ID, tile)
		}
	case *Apple:
		state.PlaySoundEffect("Eating")
		s.Fett += o.Nutrition
		o.IsConsumed = true
	case *Item:
		state.PlaySoundEffect("Item")
		s.HeldItem = o.ItemType
		o.IsConsumed = true
	case *BotSnake:
	case *FartEntity:
	case *BulletEntity:
		// Do nothing
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
