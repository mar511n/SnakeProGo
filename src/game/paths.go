package game

import (
	"image"
	"os"
	"path/filepath"
	"strings"

	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	// BaseSystemPath is the absolute path for the game data
	// This should be set via command line argument, defaulting to $HOME/snakeprogo
	BaseSystemPath = ""

	// ConfigDir is the relative path for configuration files (relative to BaseSystemPath)
	ConfigDir = "config"

	// ResDir is the relative path for game assets (relative to BaseSystemPath)
	ResDir = "res"

	// AssetsDir is the relative path for game assets (relative to ResDir)
	AssetsDir = "assets"

	// category name for snake tiles in the resource manager is "SnakeTilesBase"+[num]/"Bot" with num = 0,1,2,... for different snake colors or the bot snake
	SnakeTilesBaseDir      = "Snake"
	SnakeTileBodypartNames = []string{"SnkB", "SnkL", "SnkR", "SnkH", "SnkT"}

	// category name for items
	ItemCategoryName   = "Items"
	ItemShotBulletName = "shot_bullet"
	ItemShotTrailName  = "shot_stripes"
	ItemFartCloudName  = "fart_area"

	// category name for Food
	FoodCategoryName = "Food"
	AppleFileName    = "Apple"
)

// handles resource loading and caching. Loads all assets at the beginning of the game and keeps them in memory.
// The folder structure is represented as a map of maps for images.
// Sounds files are all stored within the "Sounds" directory. Subdirectories are treated as different sounds of the same type, which are randomly chosen when requested.
type ResourceManager struct {
	Images map[string]map[string]*ebiten.Image // category -> name -> image
	Sounds map[string][]SoundData              // name -> list of file paths for sound files
	Icons  []image.Image                       // list of icons
}

func (rm *ResourceManager) RandomSound(name string) (SoundData, bool) {
	soundList, exists := rm.Sounds[name]
	if !exists || len(soundList) == 0 {
		return nil, false
	}
	randomIndex := RandomSource.Intn(len(soundList))
	return soundList[randomIndex], true
}

func (rm *ResourceManager) LoadAssets(assetspath, images, sounds, icons string) {
	rm.Images = make(map[string]map[string]*ebiten.Image)
	rm.Sounds = make(map[string][]SoundData)
	rm.Icons = make([]image.Image, 0)

	// Load Images
	imagesPath := filepath.Join(assetspath, images)
	imageEntries, err := os.ReadDir(imagesPath)
	if err != nil {
		LogError("Error reading images directory '%s': %v", imagesPath, err)
	} else {
		for _, entry := range imageEntries {
			if entry.IsDir() {
				category := entry.Name()
				rm.Images[category] = make(map[string]*ebiten.Image)

				categoryPath := filepath.Join(imagesPath, category)
				files, err := os.ReadDir(categoryPath)
				if err != nil {
					LogError("Error reading image category '%s': %v", categoryPath, err)
					continue
				}

				for _, file := range files {
					if file.IsDir() {
						continue
					}

					ext := strings.ToLower(filepath.Ext(file.Name()))
					if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
						continue
					}

					fullPath := filepath.Join(categoryPath, file.Name())
					img, _, err := ebitenutil.NewImageFromFile(fullPath)
					if err != nil {
						LogError("Error loading image '%s': %v", fullPath, err)
						continue
					}

					name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
					rm.Images[category][name] = img
					LogInfo("Loaded image '%s' in category '%s'", name, category)
				}
			}
		}
	}

	// Load Sounds
	soundsPath := filepath.Join(assetspath, sounds)
	soundEntries, err := os.ReadDir(soundsPath)
	if err != nil {
		LogError("Error reading sounds directory '%s': %v", soundsPath, err)
	} else {
		for _, entry := range soundEntries {
			if entry.IsDir() {
				name := entry.Name()
				var soundList []SoundData

				namePath := filepath.Join(soundsPath, name)
				files, err := os.ReadDir(namePath)
				if err != nil {
					LogError("Error reading sound '%s': %v", namePath, err)
					continue
				}

				for _, file := range files {
					if file.IsDir() {
						continue
					}
					ext := strings.ToLower(filepath.Ext(file.Name()))
					if ext != ".wav" {
						continue
					}

					fullPath := filepath.Join(namePath, file.Name())
					soundList = append(soundList, SoundDataFromFile(fullPath))
				}
				rm.Sounds[name] = soundList
				LogInfo("Loaded sound '%s' with %d variations", name, len(soundList))
			} else {
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if ext != ".wav" {
					continue
				}

				name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
				fullPath := filepath.Join(soundsPath, entry.Name())
				rm.Sounds[name] = []SoundData{SoundDataFromFile(fullPath)}
				LogInfo("Loaded sound '%s'", name)
			}
		}
	}

	// Load Icons
	iconsPath := filepath.Join(assetspath, icons)
	iconEntries, err := os.ReadDir(iconsPath)
	if err != nil {
		LogError("Error reading icons directory '%s': %v", iconsPath, err)
	} else {
		for _, entry := range iconEntries {
			if entry.IsDir() {
				continue
			}

			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
				continue
			}

			fullPath := filepath.Join(iconsPath, entry.Name())
			_, img, err := ebitenutil.NewImageFromFile(fullPath)
			if err != nil {
				LogError("Error loading icon '%s': %v", fullPath, err)
				continue
			}

			rm.Icons = append(rm.Icons, img)
			LogInfo("Loaded icon '%s'", entry.Name())
		}
	}
}
