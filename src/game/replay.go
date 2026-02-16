package game

import (
	"fmt"
	"hash/fnv"
	"path"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var adjectives = []string{
	"epic", "incredible", "slimy", "wild", "furious", "sneaky", "giant", "tiny", "hungry", "angry",
	"green", "dangerous", "fast", "ancient", "mystic", "dark", "golden", "silent", "brave", "cruel",
	"crazy", "insane", "bloody", "chaotic", "peaceful", "intense", "savage", "fearsome", "magical", "venomous",
	"eldritch", "stygian", "inexorable", "voracious", "pernicious", "insidious", "serried", "labyrinthine", "arcane",
	"primordial", "spectral", "phantasmal", "cryptic", "ominous", "malevolent", "treacherous", "vengeful", "abyssal",
	"lurid", "hallowed", "unseen", "shadowy", "ethereal", "tempestuous", "perfidious", "obsidian", "emerald", "crimson",
}

var events = []string{
	"battle_of", "clash_of", "dance_of", "escape_of", "hunt_of", "journey_of", "legend_of", "mystery_of",
	"rise_of", "fall_of", "attack_of", "revenge_of", "duel_of", "invasion_of", "panic_of", "feast_of",
	"chase_of", "race_of", "doom_of", "awakening_of",
}

var participants = []string{
	"anacondas", "pythons", "vipers", "cobras", "worms", "noodles", "serpents", "reptiles", "dragons", "lizards",
	"boas", "rattlers", "mambas", "asps", "constrictors", "hydras", "basilisks", "sidewinders", "sneks", "predators",
}

type ReplaySession struct {
	currTick      int
	lastTickIdx   int
	escapePressed bool
	State         *GameState
	History       *HistoryData
	Paused        bool
	rm            *ResourceManager
}

func (r *ReplaySession) GenerateReplayVideoFilename() string {
	// Generate a unique hash based on history data
	hasher := fnv.New32a()
	hasher.Write(r.History.InitData)
	for _, v := range r.History.VarData {
		hasher.Write(v)
	}
	hash := hasher.Sum32()

	// Pick words based on the hash
	// Use the hash to seed a simple generator
	lAdj := uint32(len(adjectives))
	lEvt := uint32(len(events))
	lPar := uint32(len(participants))

	idx1 := (hash) % lAdj
	idx2 := (hash / lAdj) % lEvt
	idx3 := (hash / (lAdj * lEvt)) % lAdj
	idx4 := (hash / (lAdj * lEvt * lAdj)) % lPar

	return fmt.Sprintf("%s_%s_%s_%s.mp4", adjectives[idx1], events[idx2], adjectives[idx3], participants[idx4])
}

func (r *ReplaySession) RenderAndSaveVideo() {
	filename := r.GenerateReplayVideoFilename()
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
