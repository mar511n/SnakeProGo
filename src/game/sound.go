package game

import (
	"bytes"
	"os"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

type SoundData []byte

func SoundDataFromFile(filepath string) SoundData {
	data, err := os.ReadFile(filepath)
	if err != nil {
		LogError("Failed to open sound file %s: %v", filepath, err)
		return nil
	}
	return SoundData(data)
}

func (s SoundData) GetPlayer() *audio.Player {
	if AudioContext == nil {
		LogError("Audio context not initialized, cannot play sound")
		return nil
	}
	buffer := bytes.NewReader(s)
	stream, err := wav.DecodeF32(buffer)
	if err != nil {
		LogError("Failed to decode sound data: %v", err)
		return nil
	}
	p, err := AudioContext.NewPlayerF32(stream)
	if err != nil {
		LogError("Failed to create audio player: %v", err)
		return nil
	}
	return p
}

var AudioContext *audio.Context
var SoundResources *ResourceManager

func InitSound(rm *ResourceManager) {
	AudioContext = audio.NewContext(GConfig.AudioSampleRate)
	SoundResources = rm
	LogInfo("Audio context initialized with sample rate %d", GConfig.AudioSampleRate)
}

func PlaySound(name string) {
	if SoundResources == nil {
		LogError("Sound resources not initialized, cannot play win sound")
		return
	}
	if soundData, exists := SoundResources.RandomSound(name); exists {
		player := soundData.GetPlayer()
		if player != nil {
			player.Play()
		}
	} else {
		LogWarning("%v sound not found in resources", name)
	}
}
