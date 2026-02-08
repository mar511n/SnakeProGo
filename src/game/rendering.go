package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func Render(state *GameState, screen *ebiten.Image) {
	antialias := true
	tileSize := GConfig.ScreenHeight / state.Map.Collider.Height
	if tileSize > GConfig.ScreenWidth/state.Map.Collider.Width {
		tileSize = GConfig.ScreenWidth / state.Map.Collider.Width
	}
	toScreen := func(x, y int) (xOut, yOut int) {
		return x * tileSize, y * tileSize
	}
	// Render map
	for x, row := range state.Map.Tiles {
		for y, tile := range row {
			col := color.Gray{Y: 0}
			if tile.IsWall {
				col = color.Gray{Y: 200}
			}
			xp, yp := toScreen(x, y)
			vector.FillRect(screen, float32(xp), float32(yp), float32(tileSize), float32(tileSize), col, antialias)
		}
	}
	// Render snakes
	snakeColors := []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},
		{R: 0, G: 255, B: 0, A: 255},
		{R: 0, G: 0, B: 255, A: 255},
		{R: 255, G: 255, B: 0, A: 255},
		{R: 255, G: 0, B: 255, A: 255},
		{R: 0, G: 255, B: 255, A: 255},
	}
	for i, snake := range state.Players {
		col := snakeColors[i%len(snakeColors)]
		for _, pos := range snake.Body.Tiles {
			xp, yp := toScreen(pos.X, pos.Y)
			vector.FillRect(screen, float32(xp), float32(yp), float32(tileSize), float32(tileSize), col, antialias)
		}
	}
	// Render apples
	for _, apple := range state.Apples {
		col := color.RGBA{R: 255, G: 0, B: 0, A: 255}
		xp, yp := toScreen(apple.Collider.Tiles[0].X, apple.Collider.Tiles[0].Y)
		vector.FillCircle(screen, float32(xp)+float32(tileSize)/2, float32(yp)+float32(tileSize)/2, float32(tileSize)/2, col, antialias)
	}
	// Render items
	for _, item := range state.Items {
		col := color.RGBA{R: 0, G: 255, B: 0, A: 255}
		xp, yp := toScreen(item.Collider.Tiles[0].X, item.Collider.Tiles[0].Y)
		vector.StrokeCircle(screen, float32(xp)+float32(tileSize)/2, float32(yp)+float32(tileSize)/2, float32(tileSize)/2, float32(tileSize)/6, col, antialias)
	}
}
