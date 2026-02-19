package game

// enables differently rendered tiles and collision properties
type Tile struct {
	Name     string
	IsWall   bool
	IsSpawn  bool
	SpawnDir Vec2i
}

// implements Collidable
type MapData struct {
	Tiles       [][]Tile
	Collider    *CollisionMap // Optimised collision map (contains width, height)
	SpawnPoints []Vec2i
	SpawnDirs   []Vec2i
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
		Width:     int16(len(m.Tiles)),
		Height:    int16(len(m.Tiles[0])),
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
				m.SpawnPoints = append(m.SpawnPoints, Vec2i{X: int16(x), Y: int16(y)})
				m.SpawnDirs = append(m.SpawnDirs, tile.SpawnDir)
			}
		}
	}
}

func (m *MapData) String() string {
	if len(m.Tiles) == 0 {
		return ""
	}
	width := len(m.Tiles)
	height := len(m.Tiles[0])

	result := ""
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			tile := m.Tiles[x][y]
			switch {
			case tile.IsWall:
				result += "#"
			case tile.IsSpawn:
				switch tile.SpawnDir {
				case DirLeft:
					result += "L"
				case DirRight:
					result += "R"
				case DirUp:
					result += "U"
				case DirDown:
					result += "D"
				default:
					result += "S"
				}
			default:
				result += "."
			}
		}
		result += "\n"
	}
	return result
}

// reads a simple tilemap from a string, where each character corresponds to a tile type
func NewMapFromString(s string) *MapData {
	lines := []rune(s)
	var rows [][]Tile
	var currentRow []Tile
	for _, char := range lines {
		switch char {
		case '\n':
			if len(currentRow) > 0 {
				rows = append(rows, currentRow)
				currentRow = []Tile{}
			}
		case '#':
			currentRow = append(currentRow, Tile{Name: "wall", IsWall: true})
		case '.':
			currentRow = append(currentRow, Tile{Name: "empty"})
		case 'L':
			currentRow = append(currentRow, Tile{Name: "spawn", IsSpawn: true, SpawnDir: DirLeft})
		case 'R':
			currentRow = append(currentRow, Tile{Name: "spawn", IsSpawn: true, SpawnDir: DirRight})
		case 'U':
			currentRow = append(currentRow, Tile{Name: "spawn", IsSpawn: true, SpawnDir: DirUp})
		case 'D':
			currentRow = append(currentRow, Tile{Name: "spawn", IsSpawn: true, SpawnDir: DirDown})
		case 'S':
			currentRow = append(currentRow, Tile{Name: "spawn", IsSpawn: true})
		default:
			currentRow = append(currentRow, Tile{Name: "empty"})
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	if len(rows) == 0 {
		return &MapData{Tiles: [][]Tile{}}
	}

	height := len(rows)
	width := len(rows[0])

	tiles := make([][]Tile, width)
	for x := range tiles {
		tiles[x] = make([]Tile, height)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x < len(rows[y]) {
				tiles[x][y] = rows[y][x]
			} else {
				tiles[x][y] = Tile{Name: "empty"}
			}
		}
	}

	mapData := &MapData{Tiles: tiles}
	mapData.BuildCache()
	return mapData
}
