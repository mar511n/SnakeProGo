package game

type SpeedBoostEffect struct {
	Duration   int
	Multiplier float64
}

func (s *SpeedBoostEffect) GetType() StatusEffectType {
	return StatusEffectSpeedBoost
}

func (s *SpeedBoostEffect) Update(state *GameState) {
	s.Duration--
}

func (s *SpeedBoostEffect) IsExpired() bool {
	return s.Duration <= 0
}
