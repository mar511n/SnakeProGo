package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type ReplaySession struct {
	currTick      int
	lastTickIdx   int
	escapePressed bool
	State         *GameState
	History       *HistoryData
	Paused        bool
}

func (r *ReplaySession) Seek(tick int) {
	if tick < 0 {
		tick = int(r.History.Ticks[0]) + 1
	}
	if tick > int(r.History.Ticks[len(r.History.Ticks)-1]) {
		tick = int(r.History.Ticks[len(r.History.Ticks)-1]) - 1
	}
	r.currTick = tick
	r.lastTickIdx = -100
}

func (r *ReplaySession) IsFinished() bool {
	if r.escapePressed {
		r.escapePressed = false
		return true
	}
	return false
	//return r.currTick >= int(r.History.Ticks[len(r.History.Ticks)-1])
}

func (r *ReplaySession) Update() {
	statechange := false
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		r.Seek(r.currTick - GConfig.TPS*3)
		statechange = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		r.Seek(r.currTick + GConfig.TPS*3)
		statechange = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		r.Paused = !r.Paused
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		r.escapePressed = true
	}
	if !r.Paused && r.currTick < int(r.History.Ticks[len(r.History.Ticks)-1])-1 {
		r.currTick++
	}
	for i, t := range r.History.Ticks {
		if t > uint64(r.currTick) {
			if i-1 != r.lastTickIdx {
				r.lastTickIdx = i - 1
				statechange = true
			}
			break
		}
	}

	if statechange {
		err := r.History.ReconstructState(r.currTick, r.State)
		if err != nil {
			LogError("Failed to reconstruct game state for replay at tick %d: %v", r.currTick, err)
		}
		for _, event := range r.State.Events {
			if event.Type == GameEventSound {
				PlaySound(event.Payload.(string), GConfig.SfxVolume)
			}
		}
	}
}

func NewReplaySession(history *HistoryData) *ReplaySession {
	r := &ReplaySession{
		currTick:    0,
		lastTickIdx: 0,
		State:       &GameState{},
		History:     history,
		Paused:      false,
	}
	err := r.History.FullReconstructState(-1, r.State)
	if err != nil {
		LogError("Failed to reconstruct initial game state for replay: %v", err)
	}
	return r
}
