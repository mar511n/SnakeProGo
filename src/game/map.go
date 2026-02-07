package game

// enables differently rendered tiles and collision properties
type Tile struct {
	Name    string
	IsWall  bool
	IsSpawn bool
}

// implements Collidable
type MapData struct {
	Tiles       [][]Tile
	Collider    *CollisionMap // Optimised collision map (contains width, height)
	SpawnPoints []Vec2i
}

func (m *MapData) OnCollision(other Collidable, tile Vec2i, state *GameState) {}
func (m *MapData) OwnLayers() CollisionMask                                   { return LayerWall }
func (m *MapData) ScanLayers() CollisionMask                                  { return LayerNone }
func (m *MapData) GetCollider() CollisionObject                               { return m.Collider }
func (m *MapData) GetOwner() interface{}                                      { return m }
func (m *MapData) CanSelfCollide() bool                                       { return false }

// BuildCache populates Collider and SpawnPoints from the Tiles grid
func (m *MapData) BuildCache() {
	m.Collider = &CollisionMap{
		UseBounds: true,
		P0:        Vec2i{X: 0, Y: 0},
		Width:     len(m.Tiles),
		Height:    len(m.Tiles[0]),
		Occupied:  make([][]bool, len(m.Tiles)),
	}
	for x := range m.Tiles {
		m.Collider.Occupied[x] = make([]bool, len(m.Tiles[x]))
		for y := range m.Tiles[x] {
			tile := m.Tiles[x][y]
			if tile.IsWall {
				m.Collider.Occupied[x][y] = true
			}
			if tile.IsSpawn {
				m.SpawnPoints = append(m.SpawnPoints, Vec2i{X: x, Y: y})
			}
		}
	}
}

func (m *MapData) String() string {
	result := ""
	for _, row := range m.Tiles {
		for _, tile := range row {
			switch {
			case tile.IsWall:
				result += "#"
			case tile.IsSpawn:
				result += "S"
			default:
				result += "."
			}
		}
		result += "\n"
	}
	return result
}

// reads a simple tilemap from a string, where each character corresponds to a tile type (e.g. '#' for wall, '.' for empty, 'S' for spawn, '\n' for new row)
func NewMapFromString(s string) *MapData {
	lines := []rune(s)
	var tiles [][]Tile
	var currentRow []Tile
	for _, char := range lines {
		switch char {
		case '\n':
			if len(currentRow) > 0 {
				tiles = append(tiles, currentRow)
				currentRow = []Tile{}
			}
		case '#':
			currentRow = append(currentRow, Tile{Name: "wall", IsWall: true})
		case '.':
			currentRow = append(currentRow, Tile{Name: "empty"})
		case 'S':
			currentRow = append(currentRow, Tile{Name: "spawn", IsSpawn: true})
		default:
			currentRow = append(currentRow, Tile{Name: "empty"})
		}
	}
	if len(currentRow) > 0 {
		tiles = append(tiles, currentRow)
	}

	mapData := &MapData{Tiles: tiles}
	mapData.BuildCache()
	return mapData
}
