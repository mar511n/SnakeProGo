package game

// BaseSnake contains the physical properties and status of a snake entity.
// Implements Collidable and Updatable.
// Update method handles movement, growth, and status effects updates.
type BaseSnake struct {
	Body           *CollisionTiles // Head is at index 0 of the Vec2i slice
	Facing         Vec2f           // Current movement direction
	NextFacing     Vec2i           // Buffered input direction
	Fett           int             // "Fett" counter for growth buffer
	StatusEffects  []StatusEffect  // Active status effects (e.g. dead, invincible, speed boost)
	ticksSinceMove int             // Counter to track movement timing based on speed
}

func (s *BaseSnake) Update(state *GameState) {
	// Handle status effects: decrement durations, remove expired effects
	speed_multiplier := 1.0
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

	// Handle movement: update Body based on:
	// 	Facing and NextFacing
	// 	growth if Fett > 0
	// 	speed boosts
	// TODO: account for ticks_per_move < 1 or decide to disallow that (i.e. cap speed multiplier and keep TPS high)
	// TODO: check if this works
	current_speed := GPConfig.SnakeSpeed * speed_multiplier
	ticks_per_move := float64(GConfig.TPS) / current_speed
	s.ticksSinceMove++
	if s.ticksSinceMove >= int(ticks_per_move) {
		s.ticksSinceMove = 0

		// Update Facing based on NextFacing if it's a valid direction change
		if s.NextFacing.X != -s.Facing.X || s.NextFacing.Y != -s.Facing.Y {
			s.Facing = Vec2f{X: float64(s.NextFacing.X), Y: float64(s.NextFacing.Y)}
		}

		// Move snake body in the direction of Facing
		new_head := Vec2i{
			X: s.Body.Points[0].X + int(s.Facing.X),
			Y: s.Body.Points[0].Y + int(s.Facing.Y),
		}
		s.Body.Points = append([]Vec2i{new_head}, s.Body.Points...) // Add new head position

		if s.Fett > 0 {
			s.Fett-- // Consume one growth unit, so don't remove tail segment
		} else {
			s.Body.Points = s.Body.Points[:len(s.Body.Points)-1] // Remove tail segment
		}
	}
}

// PlayerSnake represents a player-controlled snake.
type PlayerSnake struct {
	*BaseSnake
	ID       int
	Config   *PlayerConfig // Reference to existing PlayerConfig struct (name, keys, stats)
	HeldItem ItemType      // Currently held item (ItemNone if empty)
}
