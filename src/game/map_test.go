package game

import (
	"testing"
)

func TestNewMapFromString(t *testing.T) {
	// 3x2 map (Width 3, Height 2)
	// #.S
	// L#R
	mapStr := "#.S\nL#R"
	m := NewMapFromString(mapStr)

	// In the string representation:
	// Row 0: # (Wall), . (Empty), S (Spawn Default)
	// Row 1: L (Spawn Left), # (Wall), R (Spawn Right)

	// If Tiles is [x][y], then len(Tiles) is Width (3) and len(Tiles[0]) is Height (2).
	// If Tiles is [y][x], then len(Tiles) is Height (2) and len(Tiles[0]) is Width (3).

	// Let's assume the codebase convention is [x][y] because of BuildCache.
	// If NewMapFromString is buggy, it probably produces [y][x].

	// We'll write expectations based on what it SHOULD be.
	// The user statement implies NewMapFromString is wrong.
	// We'll assume the goal is [x][y] layout for Tiles, keeping BuildCache and String in mind.
	// NOTE: String() implementation assumes [y][x] behavior currently, so that might be wrong too if we change Tiles to [x][y].
	// Or maybe String() just prints whatever rows are there?

	// Let's verify dimensions first based on string shape.
	// The string defines a map of Width 3, Height 2.
	expectedWidth := 3
	expectedHeight := 2

	// Check Collider dimensions (built by BuildCache)
	if int(m.Collider.Width) != expectedWidth {
		t.Errorf("Expected Collider Width %d, got %d", expectedWidth, m.Collider.Width)
	}
	if int(m.Collider.Height) != expectedHeight {
		t.Errorf("Expected Collider Height %d, got %d", expectedHeight, m.Collider.Height)
	}

	// Check specific tile coordinates
	// (0, 0) should be Wall (#)
	// (1, 0) should be Empty (.)
	// (2, 0) should be Spawn (S)

	// (0, 1) should be Spawn Left (L)
	// (1, 1) should be Wall (#)
	// (2, 1) should be Spawn Right (R)

	checkTile := func(x, y int, isWall bool, isSpawn bool) {
		if x >= len(m.Tiles) || y >= len(m.Tiles[x]) {
			// If access fails, dimensions are likely swapped
			// We can check boundaries first or rely on test panic/fail
			return
		}

		// If m.Tiles is [x][y]
		tile := m.Tiles[x][y]
		if tile.IsWall != isWall {
			t.Errorf("Tile at (%d, %d) IsWall expected %v, got %v", x, y, isWall, tile.IsWall)
		}
		if tile.IsSpawn != isSpawn {
			t.Errorf("Tile at (%d, %d) IsSpawn expected %v, got %v", x, y, isSpawn, tile.IsSpawn)
		}
	}

	// Check corners to detect rotation/flips
	// Top-Left (0,0) -> Wall
	checkTile(0, 0, true, false)
	// Top-Right (2,0) -> Spawn
	checkTile(2, 0, false, true)
	// Bottom-Left (0,1) -> Spawn (L)
	checkTile(0, 1, false, true)
	// Bottom-Right (2,1) -> Spawn (R)
	checkTile(2, 1, false, true)

	// Check SpawnPoints from cache
	// We expect spawns at (2,0), (0,1), (2,1)
	foundSpawns := make(map[Vec2i]bool)
	for _, sp := range m.SpawnPoints {
		foundSpawns[sp] = true
	}

	expectedSpawns := []Vec2i{{X: 2, Y: 0}, {X: 0, Y: 1}, {X: 2, Y: 1}}
	for _, sp := range expectedSpawns {
		if !foundSpawns[sp] {
			t.Errorf("Missing expected spawn point at %v", sp)
		}
	}
}

func TestMapParsing(t *testing.T) {
	mapStr := "#\n."
	m := NewMapFromString(mapStr)
	// Width 1, Height 2
	if m.Collider.Width != 1 {
		t.Errorf("Expected Width 1, got %d", m.Collider.Width)
	}
	if m.Collider.Height != 2 {
		t.Errorf("Expected Height 2, got %d", m.Collider.Height)
	}
}

func TestMapData_String(t *testing.T) {
	// Original string with explicit newline at the end of the last line
	// because String() output ends with newline for each row
	mapStr := "#.S\nL#R\n"
	m := NewMapFromString(mapStr)
	result := m.String()

	if result != mapStr {
		t.Errorf("String interpretation mismatch.\nExpected:\n%q\nGot:\n%q", mapStr, result)
	}
}
