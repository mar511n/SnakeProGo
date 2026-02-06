package game

import (
	"fmt"
)

var (
	DirUp    = Vec2i{0, -1}
	DirDown  = Vec2i{0, 1}
	DirLeft  = Vec2i{-1, 0}
	DirRight = Vec2i{1, 0}
)

// Integer vector math
type Vec2i struct {
	X, Y int
}

func (v Vec2i) Add(other Vec2i) Vec2i {
	return Vec2i{v.X + other.X, v.Y + other.Y}
}

func (v Vec2i) Sub(other Vec2i) Vec2i {
	return Vec2i{v.X - other.X, v.Y - other.Y}
}

func (v Vec2i) Mul(scalar int) Vec2i {
	return Vec2i{v.X * scalar, v.Y * scalar}
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
	return Vec2i{int(v.X), int(v.Y)}
}
