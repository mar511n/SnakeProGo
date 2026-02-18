package game

import (
	"fmt"
	"hash/fnv"
	"path"

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
	rm            *ResourceManager
}

func (r *ReplaySession) GenerateHash() uint32 {
	// Generate a unique hash based on history data
	hasher := fnv.New32a()
	hasher.Write(r.History.InitData)
	for _, v := range r.History.VarData {
		hasher.Write(v)
	}
	hash := hasher.Sum32()
	return hash
}

func (r *ReplaySession) RenderAndSaveVideo() {
	adj1, evt, adj2, par := RandomizeFilenamePartsSimple(r.GenerateHash())
	fname1 := fmt.Sprintf("%s %s %s %s", adj1, evt, adj2, par)
	filename := GenerateFilenameForReplay(fname1)
	//filename := RandomizeFilenameSimple()
	filename = path.Join(BaseSystemPath, filename)
	LogInfo("Rendering replay video to %s...", filename)
	err := r.History.RenderToVideo(filename, r.rm)
	if err != nil {
		LogError("Failed to render replay video: %v", err)
	} else {
		LogInfo("Replay video saved as %s", filename)
	}
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
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		r.RenderAndSaveVideo()
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

func NewReplaySession(history *HistoryData, rm *ResourceManager) *ReplaySession {
	r := &ReplaySession{
		currTick:    0,
		lastTickIdx: 0,
		State:       &GameState{},
		History:     history,
		Paused:      false,
		rm:          rm,
	}
	err := r.History.FullReconstructState(-1, r.State)
	if err != nil {
		LogError("Failed to reconstruct initial game state for replay: %v", err)
	}
	return r
}
