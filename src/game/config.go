package game

import (
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pelletier/go-toml/v2"
)

type GlobalConfig struct {
	AssetsDir    string `toml:"assets_dir"`    // path to assets directory
	ScreenWidth  int    `toml:"screen_width"`  // window width in pixels
	ScreenHeight int    `toml:"screen_height"` // window height in pixels
	Fullscreen   bool   `toml:"fullscreen"`    // start in fullscreen mode
	MasterVolume int    `toml:"master_volume"` // overall volume percentage
	MusicVolume  int    `toml:"music_volume"`  // music volume percentage
	SfxVolume    int    `toml:"sfx_volume"`    // sound effects volume percentage
	DebugLevel   int    `toml:"debug_level"`   // logging level (0=Error, 1=Warning, 2=Info)
	TPS          int    `toml:"tps"`           // target ticks per second
	Vsync        bool   `toml:"vsync"`         // enable vsync
}

type GameplayConfig struct {
	StartSnakeLength    int     `toml:"start_snake_length"`     // in segments
	SnakeSpeed          int     `toml:"snake_speed"`            // segments/second
	MapPath             string  `toml:"map_path"`               // path to default map file
	AppleCount          int     `toml:"apple_count"`            // number of apples on map
	AppleNutrition      int     `toml:"apple_nutrition"`        // length increase per apple
	AppleRotTime        int     `toml:"apple_rot_time"`         // time in seconds before apple rots
	GhostAppleDamage    int     `toml:"ghost_apple_damage"`     // damage per second from ghost apples
	ItemCount           int     `toml:"item_count"`             // number of items on map
	SnakeSurvivalLength int     `toml:"snake_survival_length"`  // minimum length to survive (i.e. if hit by a bullet, the snake dies if length <= this after the bullet cuts of segments)
	BulletSpeed         float64 `toml:"bullet_speed"`           // segments/second
	BulletRange         float64 `toml:"bullet_range"`           // in segments
	SpeedMultiplier     float64 `toml:"speed_multiplier"`       // speed boost multiplier
	SpeedDuration       int     `toml:"speed_duration"`         // duration of speed boost in seconds
	BotSpeed            float64 `toml:"bot_speed"`              // segments/second
	BotLength           int     `toml:"bot_length"`             // in segments
	BotDuration         int     `toml:"bot_duration"`           // duration of bot item in seconds
	FartDuration        int     `toml:"fart_duration"`          // duration of fart item in seconds
	FartSize            int     `toml:"fart_size"`              // size of fart area in segments
	FartDamagePerSecond float64 `toml:"fart_damage_per_second"` // lost segments per second inside fart area
}

type PlayerConfig struct {
	Name   string            `toml:"name"`    // player name
	KeyMap map[string]string `toml:"key_map"` // keys for [up/down/left/right] and [turn_left/turn_right] and use_item
	Stats  map[string]int    `toml:"stats"`   // player statistics like games played, apples eaten, etc.
}

var (
	GConfig  GlobalConfig
	GPConfig GameplayConfig
	PConfigs = make(map[string]PlayerConfig)

	ConfigLoaded = false
)

func LoadConfigs() {
	fullConfigDir := filepath.Join(BaseSystemPath, ConfigDir)

	if _, err := os.Stat(fullConfigDir); os.IsNotExist(err) {
		os.MkdirAll(fullConfigDir, 0755)
	}

	loadGlobalConfig()
	ConfigLoaded = true
	processGlobalConfigs()
	loadGameplayConfig()
}

func loadGlobalConfig() {
	path := filepath.Join(BaseSystemPath, ConfigDir, "config.toml")
	data, err := os.ReadFile(path)
	if err == nil {
		if err := toml.Unmarshal(data, &GConfig); err != nil {
			FatalError("Error parsing global config: %v", err)
		} else {
			LogInfo("Loaded global config.")
		}
	} else {
		LogWarning("Global config not found, using defaults.")
		GConfig = GlobalConfig{
			AssetsDir:    filepath.Join(BaseSystemPath, AssetsDir),
			ScreenWidth:  1280,
			ScreenHeight: 720,
			Fullscreen:   false,
			MasterVolume: 100,
			MusicVolume:  80,
			SfxVolume:    100,
			DebugLevel:   2,
			TPS:          60,
			Vsync:        true,
		}

		saveConfig(path, GConfig)
	}
}

func processGlobalConfigs() {
	if GConfig.AssetsDir != "" {
		AssetsDir = GConfig.AssetsDir
	}
	ebiten.SetWindowSize(GConfig.ScreenWidth, GConfig.ScreenHeight)
	ebiten.SetFullscreen(GConfig.Fullscreen)
	ebiten.SetTPS(GConfig.TPS)
	ebiten.SetVsyncEnabled(GConfig.Vsync)
}

func loadGameplayConfig() {
	path := filepath.Join(BaseSystemPath, ConfigDir, "gameplay.toml")
	data, err := os.ReadFile(path)
	if err == nil {
		if err := toml.Unmarshal(data, &GPConfig); err != nil {
			FatalError("Error parsing gameplay config: %v", err)
		} else {
			LogInfo("Loaded gameplay config.")
		}
	} else {
		LogWarning("Gameplay config not found, using defaults.")
		GPConfig = GameplayConfig{
			StartSnakeLength:    3,
			SnakeSpeed:          10,
			MapPath:             filepath.Join(BaseSystemPath, AssetsDir, "maps/default.json"),
			AppleCount:          5,
			AppleNutrition:      1,
			AppleRotTime:        300,
			GhostAppleDamage:    1,
			ItemCount:           2,
			SnakeSurvivalLength: 2,
			BulletSpeed:         0.5,
			BulletRange:         10.0,
			SpeedMultiplier:     1.5,
			SpeedDuration:       300,
			BotSpeed:            0.3,
			BotLength:           5,
			BotDuration:         600,
			FartDuration:        300,
			FartSize:            3,
			FartDamagePerSecond: 1.0,
		}

		saveConfig(path, GPConfig)
	}
}

func GetPlayerConfig(name string) PlayerConfig {
	if cfg, ok := PConfigs[name]; ok {
		return cfg
	}

	userConfigDir := filepath.Join(BaseSystemPath, ConfigDir, "userconfig")
	os.MkdirAll(userConfigDir, 0755)
	path := filepath.Join(userConfigDir, name+".toml")

	var cfg PlayerConfig
	data, err := os.ReadFile(path)
	if err == nil {
		toml.Unmarshal(data, &cfg)
		LogInfo("Loaded player config for %s.", name)
	} else {
		LogWarning("Player config for %s not found, using defaults.", name)
		cfg = PlayerConfig{
			Name: name,
			KeyMap: map[string]string{
				"up":         ebiten.KeyW.String(),
				"down":       ebiten.KeyS.String(),
				"left":       ebiten.KeyA.String(),
				"right":      ebiten.KeyD.String(),
				"turn_left":  ebiten.KeyA.String(),
				"turn_right": ebiten.KeyD.String(),
				"use_item":   ebiten.KeySpace.String(),
			},
			Stats: make(map[string]int),
		}

		saveConfig(path, cfg)
	}
	PConfigs[name] = cfg
	return cfg
}

func saveConfig(path string, v interface{}) {
	data, err := toml.Marshal(v)
	if err != nil {
		LogError("Error marshaling config: %v", err)
		return
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		LogError("Error writing config file %s: %v", path, err)
	}
}
