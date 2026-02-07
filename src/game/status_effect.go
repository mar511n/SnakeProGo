package game

type StatusEffectType int

const (
	StatusEffectCustom StatusEffectType = iota
	StatusEffectInvincibility
	StatusEffectDead
	StatusEffectGhost
	StatusEffectSpeedBoost
)

type DeadEffect struct {
}

func (d *DeadEffect) Update(state *GameState) {
	// TODO: if ghost is enabled, we can handle ghost respawn logic here, otherwise this effect just serves as a marker for death
}
func (d *DeadEffect) GetType() StatusEffectType { return StatusEffectDead }
func (d *DeadEffect) IsExpired() bool           { return false }
