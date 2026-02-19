package game

import (
	"fmt"
	"math/rand"
)

var (
	DirUp    = Vec2i{0, -1}
	DirDown  = Vec2i{0, 1}
	DirLeft  = Vec2i{-1, 0}
	DirRight = Vec2i{1, 0}
)

var (
	Periodic_Is_Initialized = false
	Periodic_P0             = Vec2i{0, 0}
	Periodic_W              = int16(100)
	Periodic_H              = int16(100)
)

func InitializePeriodicBoundary(p0 Vec2i, width, height int16) {
	if Periodic_Is_Initialized && (Periodic_P0 != p0 || Periodic_W != width || Periodic_H != height) {
		LogError("Periodic boundary already initialized")
	}
	Periodic_P0 = p0
	Periodic_W = width
	Periodic_H = height
	Periodic_Is_Initialized = true
}

// Integer vector math
type Vec2i struct {
	X, Y int16
}

func (v Vec2i) MakeP() Vec2i {
	return Vec2i{
		X: (((v.X-Periodic_P0.X)%Periodic_W)+Periodic_W)%Periodic_W + Periodic_P0.X,
		Y: (((v.Y-Periodic_P0.Y)%Periodic_H)+Periodic_H)%Periodic_H + Periodic_P0.Y,
	}
}

func (v Vec2i) DiffP(other Vec2i) Vec2i {
	// use minimum image convention to find shortest vector between two points on a periodic grid
	dx := (v.X - other.X)
	dy := (v.Y - other.Y)

	dx = ((dx + Periodic_W/2) % Periodic_W) - Periodic_W/2
	// Handle negative modulo result which can happen in Go
	if dx < -Periodic_W/2 {
		dx += Periodic_W
	}

	dy = ((dy + Periodic_H/2) % Periodic_H) - Periodic_H/2
	// Handle negative modulo result
	if dy < -Periodic_H/2 {
		dy += Periodic_H
	}

	return Vec2i{X: dx, Y: dy}
}

func (v Vec2i) Add(other Vec2i) Vec2i {
	return Vec2i{v.X + other.X, v.Y + other.Y}
}

func (v Vec2i) Sub(other Vec2i) Vec2i {
	return Vec2i{v.X - other.X, v.Y - other.Y}
}

func (v Vec2i) Mul(scalar int) Vec2i {
	return Vec2i{v.X * int16(scalar), v.Y * int16(scalar)}
}

func (v Vec2i) Equals(other Vec2i) bool {
	return v.X == other.X && v.Y == other.Y
}

func (v Vec2i) Rotate90(num_times int) Vec2i {
	num_times = ((num_times % 4) + 4) % 4 // ensure it's between 0 and 3
	switch num_times {
	case 0:
		return v
	case 1:
		return Vec2i{-v.Y, v.X}
	case 2:
		return Vec2i{-v.X, -v.Y}
	case 3:
		return Vec2i{v.Y, -v.X}
	default:
		return v // should never happen
	}
}

func (v Vec2i) Orientation() int {
	if v.X == 0 {
		if v.Y > 0 {
			return 1 // up
		} else if v.Y < 0 {
			return 3 // down
		}
	} else if v.Y == 0 {
		if v.X > 0 {
			return 0 // right
		} else if v.X < 0 {
			return 2 // left
		}
	}
	return -1
}

func (v Vec2i) String() string {
	return fmt.Sprintf("(%d, %d)", v.X, v.Y)
}

func (v Vec2i) ToVec2f() Vec2f {
	return Vec2f{float64(v.X), float64(v.Y)}
}

// Floating point vector math
type Vec2f struct {
	X, Y float64
}

func (v Vec2f) Add(other Vec2f) Vec2f {
	return Vec2f{v.X + other.X, v.Y + other.Y}
}

func (v Vec2f) Sub(other Vec2f) Vec2f {
	return Vec2f{v.X - other.X, v.Y - other.Y}
}

func (v Vec2f) Mul(scalar float64) Vec2f {
	return Vec2f{v.X * scalar, v.Y * scalar}
}

func (v Vec2f) Equals(other Vec2f, epsilon float64) bool {
	d2 := (v.X-other.X)*(v.X-other.X) + (v.Y-other.Y)*(v.Y-other.Y)
	return d2 <= epsilon*epsilon
}

func (v Vec2f) String() string {
	return fmt.Sprintf("(%.2f, %.2f)", v.X, v.Y)
}

func (v Vec2f) ToVec2i() Vec2i {
	return Vec2i{int16(v.X), int16(v.Y)}
}

var uniqueIDCounter uint64 = 1

func GetUniqueID() uint64 {
	id := uniqueIDCounter
	uniqueIDCounter++
	return id
}

var RandomSource *rand.Rand

func RandomPosition(width, height int) Vec2i {
	return Vec2i{
		X: int16(RandomSource.Intn(width)),
		Y: int16(RandomSource.Intn(height)),
	}
}
