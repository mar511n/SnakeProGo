package game

import (
	"image/color"
	"math"
	"slices"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type DefaultRenderer struct {
	Rm               *ResourceManager
	Filter           ebiten.Filter
	ItemBarWidthRel  float64
	ImageTilePadding float64

	Antialias       bool
	TileSize        float64
	ItemBarTileSize float64
	ItemBarOffset   float64
	ItemBarWidth    float64
}

func NewDefaultRenderer(rm *ResourceManager) *DefaultRenderer {
	return &DefaultRenderer{
		Rm:               rm,
		Filter:           ebiten.FilterNearest,
		ItemBarWidthRel:  0.08,
		ImageTilePadding: 0.5,
	}
}

func (r *DefaultRenderer) InitRender(useAntialias bool, state *GameState) {
	r.Antialias = useAntialias
	playernum := len(state.Players)
	r.ItemBarWidth = r.ItemBarWidthRel * float64(GConfig.ScreenWidth)
	r.TileSize = float64(GConfig.ScreenHeight) / float64(state.Map.Collider.Height)
	tsw := (float64(GConfig.ScreenWidth) - r.ItemBarWidth) / float64(state.Map.Collider.Width)
	if r.TileSize > tsw {
		r.TileSize = tsw
	}

	r.ItemBarTileSize = float64(GConfig.ScreenHeight) / float64(2*playernum+1)
	if r.ItemBarWidth/2 < r.ItemBarTileSize {
		r.ItemBarTileSize = r.ItemBarWidth / 2
	}
	r.ItemBarOffset = float64(GConfig.ScreenHeight)/2 - r.ItemBarTileSize*float64(2*playernum+1)/2
}

func (r *DefaultRenderer) GetTileDrawOptions(img *ebiten.Image, tx, ty int, oritentation, tilesnumx, tilesize float64) *ebiten.DrawImageOptions {
	op := &ebiten.DrawImageOptions{}
	op.Filter = r.Filter
	scale := (2*r.ImageTilePadding + tilesnumx) * tilesize / float64(img.Bounds().Dx())

	op.GeoM.Translate(-float64(img.Bounds().Dx())/2, -float64(img.Bounds().Dy())/2)
	op.GeoM.Rotate(math.Pi / 2 * oritentation)
	op.GeoM.Translate(float64(img.Bounds().Dx())/2, float64(img.Bounds().Dy())/2)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate((float64(tx)-r.ImageTilePadding)*tilesize, (float64(ty)-r.ImageTilePadding)*tilesize)

	return op
}

func (r *DefaultRenderer) DrawSnake(screen *ebiten.Image, tiles []Vec2i, bodyparts []*ebiten.Image) {
	// assumes bodyparts to be: snake body (right->left, right->down, right->up), snake head (->up), snake tail (left->)
	// tiles[0] is the head, tiles[len(tiles)-1] is the tail
	if len(tiles) < 2 {
		LogWarning("Snake has less than 2 tiles, cannot render body")
		return
	}
	// render tail
	tail := tiles[len(tiles)-1]
	tailOrit := tiles[len(tiles)-2].Sub(tail).Orientation()
	op := r.GetTileDrawOptions(bodyparts[4], int(tail.X), int(tail.Y), float64(tailOrit), 1, r.TileSize)
	screen.DrawImage(bodyparts[4], op)
	// render body
	for i := len(tiles) - 2; i > 0; i-- {
		tile := tiles[i]
		prev := tiles[i+1]
		next := tiles[i-1]
		prevO := tile.Sub(prev).Orientation()
		nextO := next.Sub(tile).Orientation()
		bodyO := 0
		bodyIdx := 0
		if prevO == nextO {
			bodyO = prevO - 2
			bodyIdx = 0
		} else if (prevO+1)%4 == nextO {
			bodyO = prevO - 2
			bodyIdx = 2
		} else if (prevO+3)%4 == nextO {
			bodyO = prevO - 2
			bodyIdx = 1
		} else {
			LogWarning("Invalid snake body configuration at tile %v", tile)
			continue
		}
		op := r.GetTileDrawOptions(bodyparts[bodyIdx], int(tile.X), int(tile.Y), float64(bodyO), 1, r.TileSize)
		screen.DrawImage(bodyparts[bodyIdx], op)
	}
	// render head
	head := tiles[0]
	headOrit := (head.Sub(tiles[1]).Orientation() + 1) % 4
	op = r.GetTileDrawOptions(bodyparts[3], int(head.X), int(head.Y), float64(headOrit), 1, r.TileSize)
	screen.DrawImage(bodyparts[3], op)
}

func (r *DefaultRenderer) Render(state *GameState, screen *ebiten.Image) {
	screen.Fill(color.Gray{Y: 50})

	// Render map
	for x, row := range state.Map.Tiles {
		for y, tile := range row {
			if tile.IsWall {
				col := color.Gray{Y: 200}
				vector.FillRect(screen, float32(x)*float32(r.TileSize), float32(y)*float32(r.TileSize), float32(r.TileSize), float32(r.TileSize), col, r.Antialias)
			}
		}
	}

	// Render apples
	for _, apple := range state.Apples {
		img, ok := r.Rm.Images[FoodCategoryName][AppleFileName]
		if !ok {
			LogError("Apple image not found in resource manager, cannot render apple")
			continue
		}
		op := r.GetTileDrawOptions(img, int(apple.Collider.Tiles[0].X), int(apple.Collider.Tiles[0].Y), 0, 1, r.TileSize)
		screen.DrawImage(img, op)
	}

	// Render items
	for _, item := range state.Items {
		img, ok := r.Rm.Images[ItemCategoryName][item.ItemType.FileName()]
		if !ok {
			LogError("Item image for item type %v not found in resource manager, cannot render item", item.ItemType)
			continue
		}
		op := r.GetTileDrawOptions(img, int(item.Collider.Tiles[0].X), int(item.Collider.Tiles[0].Y), 0, 1, r.TileSize)
		screen.DrawImage(img, op)
	}

	// Render snakes
	// sort player IDs to ensure consistent rendering order
	IDs := make([]int, 0, len(state.Players))
	for id := range state.Players {
		IDs = append(IDs, id)
	}
	slices.Sort(IDs)
	for idx, id := range IDs {
		snake := state.Players[id]
		if len(snake.Body.Tiles) < 2 {
			continue
		}
		bodyparts := make([]*ebiten.Image, 5)
		drawSnake := true
		for i := 0; i < 5; i++ {
			bodyparts[i], drawSnake = r.Rm.Images[SnakeTilesBaseDir+strconv.Itoa(idx)][SnakeTileBodypartNames[i]]
			if !drawSnake {
				LogError("Snake body part image '%v' for snake '%v' not found in resource manager, cannot render snake", SnakeTileBodypartNames[i], SnakeTilesBaseDir+strconv.Itoa(idx))
				continue
			}
		}
		if !drawSnake {
			continue
		}
		r.DrawSnake(screen, snake.Body.Tiles, bodyparts)
	}

	// Render UI
	vector.FillRect(screen, float32(GConfig.ScreenWidth)-float32(r.ItemBarWidth), 0, float32(r.ItemBarWidth), float32(GConfig.ScreenHeight), color.Gray{Y: 80}, r.Antialias)
	for idx, id := range IDs {
		snake := state.Players[id]
		// render snake head as icon for item bar
		img, ok := r.Rm.Images[SnakeTilesBaseDir+strconv.Itoa(idx)][SnakeTileBodypartNames[3]]
		if !ok {
			LogError("Snake head image for player %d not found in resource manager, cannot render item bar", id)
			continue
		}
		op := r.GetTileDrawOptions(img, 0, 0, -1, 1, r.ItemBarTileSize)
		op.GeoM.Translate(float64(GConfig.ScreenWidth)-r.ItemBarTileSize, r.ItemBarOffset+float64(1+2*idx)*r.ItemBarTileSize)
		screen.DrawImage(img, op)
		// render item below snake head
		if snake.HeldItem != ItemNone {
			img, ok := r.Rm.Images[ItemCategoryName][snake.HeldItem.FileName()]
			if !ok {
				LogError("Item image for held item %v of player %d not found in resource manager, cannot render item bar", snake.HeldItem.FileName(), id)
				continue
			}
			op := r.GetTileDrawOptions(img, 0, 0, 0, 1, r.ItemBarTileSize)
			op.GeoM.Translate(float64(GConfig.ScreenWidth)-2*r.ItemBarTileSize, r.ItemBarOffset+float64(1+2*idx)*r.ItemBarTileSize)
			screen.DrawImage(img, op)
		}
	}

	if GConfig.DisplayFPS {
		ebitenutil.DebugPrintAt(screen, "FPS: "+strconv.Itoa(int(ebiten.ActualFPS())), 10, 10)
	}
	if GConfig.DisplayTPS {
		ebitenutil.DebugPrintAt(screen, "TPS: "+strconv.Itoa(int(ebiten.ActualTPS())), 10, 30)
	}
}
