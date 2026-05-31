package game

type Statistics struct {
	TotalGamesPlayed int `toml:"TotalGamesPlayed"`
	TotalWins        int `toml:"TotalWins"`
	TotalKills       int `toml:"TotalKills"`
	TotalDeaths      int `toml:"TotalDeaths"`
	TotalItemsUsed   int `toml:"TotalItemsUsed"`
}

func NewStatistics() *Statistics {
	return &Statistics{
		TotalGamesPlayed: 0,
		TotalWins:        0,
		TotalKills:       0,
		TotalDeaths:      0,
		TotalItemsUsed:   0,
	}
}

func (s *Statistics) Copy() *Statistics {
	return &Statistics{
		TotalGamesPlayed: s.TotalGamesPlayed,
		TotalWins:        s.TotalWins,
		TotalKills:       s.TotalKills,
		TotalDeaths:      s.TotalDeaths,
		TotalItemsUsed:   s.TotalItemsUsed,
	}
}
