package game

type CheckWinCondition func(state *GameState) (bool, []int) // returns (win, winnerIDs)

func CheckLastOneStanding(state *GameState) (bool, []int) {
	aliveCount := 0
	var lastAliveID int
	for id, player := range state.Players {
		if !player.IsDead() {
			aliveCount++
			lastAliveID = id
		}
	}
	if aliveCount == 0 {
		return true, []int{} // No winners, all dead
	}
	return aliveCount == 1, []int{lastAliveID}
}
