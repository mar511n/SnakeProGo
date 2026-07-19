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
	screenwidth      int
	screenheight     int
	displayTPS       bool
	displayFPS       bool
	drawperiodic     bool

	Antialias              bool
	TileSize               float64
	ItemBarTileSize        float64
	ItemBarOffset          float64
	ItemBarWidth           float64
	GameAreaX0, GameAreaY0 float64
	GameAreaW, GameAreaH   float64
}

func NewDefaultRenderer(rm *ResourceManager) *DefaultRenderer {
	return &DefaultRenderer{
		Rm:               rm,
		Filter:           ebiten.FilterNearest,
		ItemBarWidthRel:  0.08,
		ImageTilePadding: 0.5,
		screenwidth:      GConfig.ScreenWidth,
		screenheight:     GConfig.ScreenHeight,
		displayTPS:       GConfig.DisplayTPS,
		displayFPS:       GConfig.DisplayFPS,
		drawperiodic:     true,
	}
}

func (r *DefaultRenderer) InitRender(useAntialias bool, state *GameState) {
	r.Antialias = useAntialias
	playernum := len(state.Players)
	r.ItemBarWidth = r.ItemBarWidthRel * float64(r.screenwidth)
	r.TileSize = float64(r.screenheight) / float64(state.Map.Collider.Height)
	tsw := (float64(r.screenwidth) - r.ItemBarWidth) / float64(state.Map.Collider.Width)
	if r.TileSize > tsw {
		r.TileSize = tsw
	}

	r.ItemBarTileSize = float64(r.screenheight) / float64(2*playernum+1)
	if r.ItemBarWidth/2 < r.ItemBarTileSize {
		r.ItemBarTileSize = r.ItemBarWidth / 2
	}
	r.ItemBarOffset = float64(r.screenheight)/2 - r.ItemBarTileSize*float64(2*playernum+1)/2
	r.GameAreaW = r.TileSize * float64(state.Map.Collider.Width)
	r.GameAreaH = r.TileSize * float64(state.Map.Collider.Height)
	r.GameAreaX0 = (float64(r.screenwidth) - r.ItemBarWidth - r.GameAreaW) / 2
	r.GameAreaY0 = (float64(r.screenheight) - r.GameAreaH) / 2
}

func (r *DefaultRenderer) DrawImagePeriodic(screen *ebiten.Image, img *ebiten.Image, op *ebiten.DrawImageOptions) {
	// Draw the image 9 times around the rectangular area defined by p0, w, h to create a seamless periodic tiling effect
	if !r.drawperiodic {
		screen.DrawImage(img, op)
		return
	}
	for _, d := range []Vec2i{{0, 0}, DirLeft, DirRight, DirUp, DirDown} {
		opCopy := *op
		opCopy.GeoM.Translate(float64(d.X)*r.GameAreaW, float64(d.Y)*r.GameAreaH)
		screen.DrawImage(img, &opCopy)
	}
}

func (r *DefaultRenderer) GetTileDrawOptions(img *ebiten.Image, tx, ty int, oritentation, tilesnumx, tilesize, tilepadding float64, gameoffset bool) *ebiten.DrawImageOptions {
	op := &ebiten.DrawImageOptions{}
	op.Filter = r.Filter
	scale := (2*tilepadding + tilesnumx) * tilesize / float64(img.Bounds().Dx())

	op.GeoM.Translate(-float64(img.Bounds().Dx())/2, -float64(img.Bounds().Dy())/2)
	op.GeoM.Rotate(math.Pi / 2 * oritentation)
	op.GeoM.Translate(float64(img.Bounds().Dx())/2, float64(img.Bounds().Dy())/2)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate((float64(tx)-tilepadding)*tilesize, (float64(ty)-tilepadding)*tilesize)
	if gameoffset {
		op.GeoM.Translate(r.GameAreaX0, r.GameAreaY0)
	}
	return op
}

func (r *DefaultRenderer) DrawSnake(screen *ebiten.Image, tiles []Vec2i, facing Vec2i, bodyparts []*ebiten.Image, isDead bool) {
	// assumes bodyparts to be: snake body (right->left, right->down, right->up), snake head (->up), snake tail (left->)
	// tiles[0] is the head, tiles[len(tiles)-1] is the tail
	if len(tiles) == 0 {
		LogWarning("Attempted to render snake with no body tiles, skipping render")
		return
	}
	cs := &ebiten.ColorScale{}
	if isDead {
		cs.ScaleAlpha(0.5)
	}
	if len(tiles) >= 2 {
		// render tail
		tail := tiles[len(tiles)-1]
		tailOrit := tiles[len(tiles)-2].DiffP(tail).Orientation()
		op := r.GetTileDrawOptions(bodyparts[4], int(tail.X), int(tail.Y), float64(tailOrit), 1, r.TileSize, r.ImageTilePadding, true)
		op.ColorScale = *cs
		r.DrawImagePeriodic(screen, bodyparts[4], op)
	}
	if len(tiles) > 2 {
		// render body
		for i := len(tiles) - 2; i > 0; i-- {
			tile := tiles[i]
			prev := tiles[i+1]
			next := tiles[i-1]
			prevO := tile.DiffP(prev).Orientation()
			nextO := next.DiffP(tile).Orientation()
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
			op := r.GetTileDrawOptions(bodyparts[bodyIdx], int(tile.X), int(tile.Y), float64(bodyO), 1, r.TileSize, r.ImageTilePadding, true)
			op.ColorScale = *cs
			r.DrawImagePeriodic(screen, bodyparts[bodyIdx], op)
		}
	}
	// render head
	head := tiles[0]
	headOrit := (facing.Orientation() + 1) % 4
	if len(tiles) > 1 {
		headOrit = (head.DiffP(tiles[1]).Orientation() + 1) % 4
	}
	op := r.GetTileDrawOptions(bodyparts[3], int(head.X), int(head.Y), float64(headOrit), 1, r.TileSize, r.ImageTilePadding, true)
	op.ColorScale = *cs
	r.DrawImagePeriodic(screen, bodyparts[3], op)
}

func (r *DefaultRenderer) Render(state *GameState, screen *ebiten.Image) {
	screen.Fill(color.Gray{Y: 0})

	// Render map
	for x, row := range state.Map.Tiles {
		for y, tile := range row {
			col := color.Gray{Y: 50}
			if tile.IsWall {
				col = color.Gray{Y: 200}
			}
			vector.FillRect(screen, float32(x)*float32(r.TileSize)+float32(r.GameAreaX0), float32(y)*float32(r.TileSize)+float32(r.GameAreaY0), float32(r.TileSize)*1.01, float32(r.TileSize)*1.01, col, r.Antialias)
		}
	}

	// Render entities that are at the bottom
	for _, entity := range state.Entities {
		switch e := entity.(type) {
		case *FartEntity:
			fart_img := r.Rm.Images[ItemCategoryName][ItemFartCloudName]
			op := r.GetTileDrawOptions(fart_img, int(e.Center.X)-GPConfig.FartSize, int(e.Center.Y)-GPConfig.FartSize, 0, float64(1+2*GPConfig.FartSize), r.TileSize, 0.0, true)
			r.DrawImagePeriodic(screen, fart_img, op)
		case *BotSnake:
			if e.IsDead() {
				continue
			}
			snake := e.BaseSnake
			bodyparts := make([]*ebiten.Image, 5)
			drawSnake := true
			for i := 0; i < 5; i++ {
				bodyparts[i], drawSnake = r.Rm.Images[SnakeTilesBaseDir+"Bot"][SnakeTileBodypartNames[i]]
				if !drawSnake {
					LogError("Snake body part image '%v' for snake '%v' not found in resource manager, cannot render snake", SnakeTileBodypartNames[i], SnakeTilesBaseDir+"Bot")
					continue
				}
			}
			if !drawSnake {
				continue
			}
			r.DrawSnake(screen, snake.Body.Tiles, snake.Facing, bodyparts, false)
		}
	}

	// Render apples
	for _, apple := range state.Apples {
		img, ok := r.Rm.Images[FoodCategoryName][AppleFileName]
		if !ok {
			LogError("Apple image not found in resource manager, cannot render apple")
			continue
		}
		op := r.GetTileDrawOptions(img, int(apple.Collider.Tiles[0].X), int(apple.Collider.Tiles[0].Y), 0, 1, r.TileSize, r.ImageTilePadding, true)
		screen.DrawImage(img, op)
	}

	// Render items
	for _, item := range state.Items {
		img, ok := r.Rm.Images[ItemCategoryName][item.ItemType.FileName()]
		if !ok {
			LogError("Item image for item type %v not found in resource manager, cannot render item", item.ItemType)
			continue
		}
		op := r.GetTileDrawOptions(img, int(item.Collider.Tiles[0].X), int(item.Collider.Tiles[0].Y), 0, 1, r.TileSize, r.ImageTilePadding, true)
		screen.DrawImage(img, op)
	}

	// Render entities above apples and items
	for _, entity := range state.Entities {
		switch e := entity.(type) {
		case *BulletEntity:
			bullet_img := r.Rm.Images[ItemCategoryName][ItemShotBulletName]
			trail_img := r.Rm.Images[ItemCategoryName][ItemShotTrailName]
			for _, pos := range e.Trail[:len(e.Trail)-1] {
				op := r.GetTileDrawOptions(trail_img, int(pos.X), int(pos.Y), float64((e.Dir.Orientation()+1)%4), 1, r.TileSize, r.ImageTilePadding, true)
				r.DrawImagePeriodic(screen, trail_img, op)
			}
			op := r.GetTileDrawOptions(bullet_img, int(e.Collider.Tiles[0].X), int(e.Collider.Tiles[0].Y), float64((e.Dir.Orientation()+1)%4), 1, r.TileSize, r.ImageTilePadding, true)
			r.DrawImagePeriodic(screen, bullet_img, op)
		}
	}

	// Render player snakes
	// sort player IDs to ensure consistent rendering order
	IDs := make([]int, 0, len(state.Players))
	for id := range state.Players {
		IDs = append(IDs, id)
	}
	slices.Sort(IDs)
	for idx, id := range IDs {
		snake := state.Players[id]
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
		r.DrawSnake(screen, snake.Body.Tiles, snake.Facing, bodyparts, snake.IsDead())
	}

	// Render UI
	vector.FillRect(screen, float32(r.screenwidth)-float32(r.ItemBarWidth), 0, float32(r.ItemBarWidth), float32(r.screenheight), color.Gray{Y: 80}, r.Antialias)
	for idx, id := range IDs {
		snake := state.Players[id]
		// render snake head as icon for item bar
		img, ok := r.Rm.Images[SnakeTilesBaseDir+strconv.Itoa(idx)][SnakeTileBodypartNames[3]]
		if !ok {
			LogError("Snake head image for player %d not found in resource manager, cannot render item bar", id)
			continue
		}
		op := r.GetTileDrawOptions(img, 0, 0, -1, 1, r.ItemBarTileSize, r.ImageTilePadding, false)
		op.GeoM.Translate(float64(r.screenwidth)-r.ItemBarTileSize, r.ItemBarOffset+float64(1+2*idx)*r.ItemBarTileSize)
		screen.DrawImage(img, op)
		// render item below snake head
		if snake.HeldItem != ItemNone {
			img, ok := r.Rm.Images[ItemCategoryName][snake.HeldItem.FileName()]
			if !ok {
				LogError("Item image for held item %v of player %d not found in resource manager, cannot render item bar", snake.HeldItem.FileName(), id)
				continue
			}
			op := r.GetTileDrawOptions(img, 0, 0, 0, 1, r.ItemBarTileSize, r.ImageTilePadding, false)
			op.GeoM.Translate(float64(r.screenwidth)-2*r.ItemBarTileSize, r.ItemBarOffset+float64(1+2*idx)*r.ItemBarTileSize)
			screen.DrawImage(img, op)
		}
	}

	if r.displayFPS {
		ebitenutil.DebugPrintAt(screen, "FPS: "+strconv.Itoa(int(ebiten.ActualFPS())), 10, 10)
	}
	if r.displayTPS {
		ebitenutil.DebugPrintAt(screen, "TPS: "+strconv.Itoa(int(ebiten.ActualTPS())), 10, 30)
	}
}
