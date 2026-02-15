package game

type StatusEffectType int

const (
	StatusEffectCustom StatusEffectType = iota
	StatusEffectInvincibility
	StatusEffectDead
	StatusEffectGhost
	StatusEffectSpeedBoost
	StatusEffectRespawning
)

func (s StatusEffectType) String() string {
	switch s {
	case StatusEffectCustom:
		return "Custom"
	case StatusEffectInvincibility:
		return "Invincibility"
	case StatusEffectDead:
		return "Dead"
	case StatusEffectGhost:
		return "Ghost"
	case StatusEffectSpeedBoost:
		return "Speed Boost"
	case StatusEffectRespawning:
		return "Respawning"
	default:
		return "Unknown"
	}
}

func NewDeadStatusEffect() *StatusEffect {
	return &StatusEffect{
		Type:     StatusEffectDead,
		duration: 1,
	}
}
func NewRespawningStatusEffect(duration float64) *StatusEffect {
	return &StatusEffect{
		Type:     StatusEffectRespawning,
		duration: int(duration * float64(GConfig.TPS)),
	}
}
func NewSpeedBoostStatusEffect(duration float64) *StatusEffect {
	return &StatusEffect{
		Type:     StatusEffectSpeedBoost,
		duration: int(duration * float64(GConfig.TPS)),
	}
}
func NewInvincibilityStatusEffect(duration float64) *StatusEffect {
	return &StatusEffect{
		Type:     StatusEffectInvincibility,
		duration: int(duration * float64(GConfig.TPS)),
	}
}

type StatusEffect struct {
	Type     StatusEffectType
	duration int
}

func (s *StatusEffect) Update(owner *BaseSnake, state *GameState, hist *HistoryData) {
	switch s.Type {
	case StatusEffectRespawning:
		s.duration--
		if s.duration == 0 {
			owner.StatusEffects = []*StatusEffect{&StatusEffect{
				Type:     StatusEffectInvincibility,
				duration: int(GPConfig.ReviveInvincibilityDuration * float64(GConfig.TPS)),
			}}
		} else if s.duration > 0 {
			owner.Facing = owner.NextFacing
		}
	case StatusEffectDead:
		s.duration = 1
		// DeadEffect doesn't have any update logic, but we still need to handle it here to prevent the default case from decrementing its duration.
	case StatusEffectSpeedBoost:
		s.duration--
	default:
		s.duration--
	}
}
func (s *StatusEffect) IsExpired() bool {
	return s.duration <= 0
}
