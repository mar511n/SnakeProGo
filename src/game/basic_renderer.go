package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type BasicRenderer struct {
	antialias   bool
	tileSize    int
	snakeColors []color.RGBA
}

func (r *BasicRenderer) toScreen(x, y int) (xOut, yOut int) {
	return x * r.tileSize, y * r.tileSize
}

func (r *BasicRenderer) InitRender(useAntialias bool, state *GameState) {
	r.antialias = useAntialias
	r.tileSize = GConfig.ScreenHeight / state.Map.Collider.Height
	if r.tileSize > GConfig.ScreenWidth/state.Map.Collider.Width {
		r.tileSize = GConfig.ScreenWidth / state.Map.Collider.Width
	}
	r.snakeColors = []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},
		{R: 0, G: 255, B: 0, A: 255},
		{R: 0, G: 0, B: 255, A: 255},
		{R: 255, G: 255, B: 0, A: 255},
		{R: 255, G: 0, B: 255, A: 255},
		{R: 0, G: 255, B: 255, A: 255},
	}
}

func (r *BasicRenderer) Render(state *GameState, screen *ebiten.Image) {
	// Render map
	for x, row := range state.Map.Tiles {
		for y, tile := range row {
			col := color.Gray{Y: 0}
			if tile.IsWall {
				col = color.Gray{Y: 200}
			}
			xp, yp := r.toScreen(x, y)
			vector.FillRect(screen, float32(xp), float32(yp), float32(r.tileSize), float32(r.tileSize), col, r.antialias)
		}
	}
	// Render snakes
	for i, snake := range state.Players {
		col := r.snakeColors[i%len(r.snakeColors)]
		for _, pos := range snake.Body.Tiles {
			xp, yp := r.toScreen(pos.X, pos.Y)
			vector.FillRect(screen, float32(xp), float32(yp), float32(r.tileSize), float32(r.tileSize), col, r.antialias)
		}
	}
	// Render apples
	for _, apple := range state.Apples {
		col := color.RGBA{R: 255, G: 0, B: 0, A: 255}
		xp, yp := r.toScreen(apple.Collider.Tiles[0].X, apple.Collider.Tiles[0].Y)
		vector.FillCircle(screen, float32(xp)+float32(r.tileSize)/2, float32(yp)+float32(r.tileSize)/2, float32(r.tileSize)/2, col, r.antialias)
	}
	// Render items
	for _, item := range state.Items {
		col := color.RGBA{R: 0, G: 255, B: 0, A: 255}
		xp, yp := r.toScreen(item.Collider.Tiles[0].X, item.Collider.Tiles[0].Y)
		vector.StrokeCircle(screen, float32(xp)+float32(r.tileSize)/2, float32(yp)+float32(r.tileSize)/2, float32(r.tileSize)/2, float32(r.tileSize)/6, col, r.antialias)
	}
}
