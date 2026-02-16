package game

import (
	"bytes"
	"errors"

	"github.com/hajimehoshi/ebiten/v2"
)

type HistoryData struct {
	VarData  [][]byte
	Ticks    []uint64
	VarSize  int
	InitData []byte
}

func (h *HistoryData) RenderToVideo(filename string, rm *ResourceManager) error {
	video := make([]*ebiten.Image, len(h.VarData))
	frameIndices := make([]int, len(h.VarData))

	state := &GameState{}
	err := state.UnmarshalAllObjects(h.InitData)
	if err != nil {
		return err
	}

	renderer := NewDefaultRenderer(rm)
	renderer.InitRender(true, state)
	renderer.displayFPS = false
	renderer.displayTPS = false

	for i, data := range h.VarData {
		video[i] = ebiten.NewImage(GConfig.ScreenWidth, GConfig.ScreenHeight)
		frameIndices[i] = int(h.Ticks[i])
		err := state.UnmarshalMutableObjects(data)
		state.Tick = h.Ticks[i]
		if err != nil {
			return err
		}
		renderer.Render(state, video[i])
	}

	return RenderVideo(video, frameIndices, GConfig.TPS, filename)
}

func (h *HistoryData) ReconstructState(tick int, state *GameState) error {
	if len(h.VarData) == 0 {
		return nil
	}
	ti := -1
	if tick > 0 {
		for i, t := range h.Ticks {
			if t >= uint64(tick) {
				ti = i
				break
			}
		}
	}
	if ti >= 0 {
		err := state.UnmarshalMutableObjects(h.VarData[ti])
		state.Tick = h.Ticks[ti]
		return err
	}
	return nil
}

func (h *HistoryData) FullReconstructState(tick int, state *GameState) error {
	if len(h.InitData) == 0 {
		return errors.New("no initial state data in history")
	}
	err := state.UnmarshalAllObjects(h.InitData)
	if err != nil {
		return err
	}
	return h.ReconstructState(tick, state)
}

func (h *HistoryData) Init(state *GameState) {
	initData, err := state.MarshalAllObjects()
	if err != nil {
		LogError("Failed to serialize initial game state for history: %v", err)
	} else {
		h.InitData = initData
	}
}

func (h *HistoryData) AddEntry(state *GameState) (updated bool) {
	updated = false
	b, err := state.MarshalMutableObjects()

	if err != nil {
		LogError("Failed to serialize game state for history: %v", err)
	} else {
		// check if b is different from the last entry in s.History to avoid storing duplicate states
		if len(h.VarData) == 0 || !bytes.Equal(b, h.VarData[len(h.VarData)-1]) {
			updated = true
			//LogInfo("Storing game state for history (tick %d, size %d), history size: %d", state.Tick, len(b), len(h.VarData))
			h.VarData = append(h.VarData, b)
			h.Ticks = append(h.Ticks, state.Tick)
			h.VarSize += len(b)
			// if history size exceeds max, remove half of the oldest entries
			if h.VarSize > GConfig.MaxHistorySize {
				removeCount := len(h.VarData) / 2
				for i := 0; i < removeCount; i++ {
					h.VarSize -= len(h.VarData[i])
				}
				h.VarData = h.VarData[removeCount:]
				h.Ticks = h.Ticks[removeCount:]
				LogInfo("History size exceeded max, removed %d oldest entries, new size %d", removeCount, h.VarSize)
			}
		}
	}
	return
}
