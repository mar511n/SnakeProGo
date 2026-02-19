package game

import (
	"testing"
)

func TestVec2i_Add(t *testing.T) {
	tests := []struct {
		name     string
		v, other Vec2i
		expected Vec2i
	}{
		{"positive numbers", Vec2i{1, 2}, Vec2i{3, 4}, Vec2i{4, 6}},
		{"mixed numbers", Vec2i{10, -5}, Vec2i{-3, 2}, Vec2i{7, -3}},
		{"negative numbers", Vec2i{-1, -1}, Vec2i{-2, -3}, Vec2i{-3, -4}},
		{"zeros", Vec2i{0, 0}, Vec2i{0, 0}, Vec2i{0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.Add(tt.other)
			if result != tt.expected {
				t.Errorf("%v.Add(%v) = %v; expected %v", tt.v, tt.other, result, tt.expected)
			}
		})
	}
}

func TestVec2i_Sub(t *testing.T) {
	tests := []struct {
		name     string
		v, other Vec2i
		expected Vec2i
	}{
		{"positive numbers", Vec2i{5, 6}, Vec2i{2, 3}, Vec2i{3, 3}},
		{"mixed numbers", Vec2i{10, -5}, Vec2i{-3, 2}, Vec2i{13, -7}},
		{"result zero", Vec2i{1, 1}, Vec2i{1, 1}, Vec2i{0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.Sub(tt.other)
			if result != tt.expected {
				t.Errorf("%v.Sub(%v) = %v; expected %v", tt.v, tt.other, result, tt.expected)
			}
		})
	}
}

func TestVec2i_Mul(t *testing.T) {
	tests := []struct {
		name     string
		v        Vec2i
		scalar   int
		expected Vec2i
	}{
		{"positive scalar", Vec2i{2, 3}, 3, Vec2i{6, 9}},
		{"negative scalar", Vec2i{2, -3}, -2, Vec2i{-4, 6}},
		{"zero scalar", Vec2i{100, 100}, 0, Vec2i{0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.Mul(tt.scalar)
			if result != tt.expected {
				t.Errorf("%v.Mul(%d) = %v; expected %v", tt.v, tt.scalar, result, tt.expected)
			}
		})
	}
}

func TestVec2i_Equals(t *testing.T) {
	tests := []struct {
		name     string
		v, other Vec2i
		expected bool
	}{
		{"equal", Vec2i{1, 2}, Vec2i{1, 2}, true},
		{"not equal x", Vec2i{1, 2}, Vec2i{2, 2}, false},
		{"not equal y", Vec2i{1, 2}, Vec2i{1, 3}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.Equals(tt.other)
			if result != tt.expected {
				t.Errorf("%v.Equals(%v) = %v; expected %v", tt.v, tt.other, result, tt.expected)
			}
		})
	}
}

func TestVec2i_Rotate90(t *testing.T) {
	v := Vec2i{1, 0}
	tests := []struct {
		name     string
		numTimes int
		expected Vec2i
	}{
		{"0 times", 0, Vec2i{1, 0}},
		{"1 time (90 deg)", 1, Vec2i{0, 1}},
		{"2 times (180 deg)", 2, Vec2i{-1, 0}},
		{"3 times (270 deg)", 3, Vec2i{0, -1}},
		{"4 times (360 deg)", 4, Vec2i{1, 0}},
		{"negative rotation", -1, Vec2i{0, -1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.Rotate90(tt.numTimes)
			if result != tt.expected {
				t.Errorf("%v.Rotate90(%d) = %v; expected %v", v, tt.numTimes, result, tt.expected)
			}
		})
	}
}

func TestVec2i_String(t *testing.T) {
	v := Vec2i{X: 10, Y: -5}
	expected := "(10, -5)"
	if got := v.String(); got != expected {
		t.Errorf("Vec2i.String() = %q, want %q", got, expected)
	}
}

func TestVec2i_ToVec2f(t *testing.T) {
	v := Vec2i{X: 1, Y: 2}
	expected := Vec2f{X: 1.0, Y: 2.0}
	got := v.ToVec2f()
	if got != expected {
		t.Errorf("Vec2i.ToVec2f() = %v, want %v", got, expected)
	}
}

func TestVec2f_Add(t *testing.T) {
	tests := []struct {
		name     string
		v, other Vec2f
		expected Vec2f
	}{
		{"simple add", Vec2f{1.5, 2.5}, Vec2f{0.5, 0.5}, Vec2f{2.0, 3.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.Add(tt.other)
			if result != tt.expected {
				t.Errorf("%v.Add(%v) = %v; expected %v", tt.v, tt.other, result, tt.expected)
			}
		})
	}
}

func TestVec2f_Sub(t *testing.T) {
	tests := []struct {
		name     string
		v, other Vec2f
		expected Vec2f
	}{
		{"simple sub", Vec2f{2.5, 3.5}, Vec2f{0.5, 1.5}, Vec2f{2.0, 2.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.Sub(tt.other)
			if result != tt.expected {
				t.Errorf("%v.Sub(%v) = %v; expected %v", tt.v, tt.other, result, tt.expected)
			}
		})
	}
}

func TestVec2f_Mul(t *testing.T) {
	tests := []struct {
		name     string
		v        Vec2f
		scalar   float64
		expected Vec2f
	}{
		{"multiplication", Vec2f{1.5, 2.0}, 2.0, Vec2f{3.0, 4.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.Mul(tt.scalar)
			if result != tt.expected {
				t.Errorf("%v.Mul(%f) = %v; expected %v", tt.v, tt.scalar, result, tt.expected)
			}
		})
	}
}

func TestVec2f_Equals(t *testing.T) {
	tests := []struct {
		name     string
		v, other Vec2f
		epsilon  float64
		expected bool
	}{
		{"exact match", Vec2f{1.0, 2.0}, Vec2f{1.0, 2.0}, 0.0001, true},
		{"within epsilon", Vec2f{1.0, 2.0}, Vec2f{1.000001, 2.0}, 0.00001, true},
		{"outside epsilon", Vec2f{1.0, 2.0}, Vec2f{1.1, 2.0}, 0.0001, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.Equals(tt.other, tt.epsilon)
			if result != tt.expected {
				t.Errorf("%v.Equals(%v, %f) = %v; expected %v", tt.v, tt.other, tt.epsilon, result, tt.expected)
			}
		})
	}
}

func TestVec2f_String(t *testing.T) {
	v := Vec2f{X: 1.234, Y: 5.678}
	expected := "(1.23, 5.68)"
	if got := v.String(); got != expected {
		t.Errorf("Vec2f.String() = %q, want %q", got, expected)
	}
}

func TestVec2f_ToVec2i(t *testing.T) {
	v := Vec2f{X: 3.9, Y: 4.1}
	expected := Vec2i{X: 3, Y: 4}
	got := v.ToVec2i()
	if got != expected {
		t.Errorf("Vec2f.ToVec2i() = %v, want %v", got, expected)
	}
}

func TestVec2i_Periodic(t *testing.T) {
	// Setup periodic boundary
	Periodic_P0 = Vec2i{0, 0}
	Periodic_W = 10
	Periodic_H = 10
	Periodic_Is_Initialized = true

	t.Run("MakeP", func(t *testing.T) {
		tests := []struct {
			name     string
			v        Vec2i
			expected Vec2i
		}{
			{"inside", Vec2i{5, 5}, Vec2i{5, 5}},
			{"overflow positive", Vec2i{12, 12}, Vec2i{2, 2}},
			{"overflow negative", Vec2i{-2, -2}, Vec2i{8, 8}},
			{"large overflow positive", Vec2i{22, 22}, Vec2i{2, 2}},
			{"large overflow negative", Vec2i{-12, -12}, Vec2i{8, 8}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := tt.v.MakeP()
				if got != tt.expected {
					t.Errorf("%v.MakeP() = %v; expected %v", tt.v, got, tt.expected)
				}
			})
		}
	})

	t.Run("DiffP", func(t *testing.T) {
		tests := []struct {
			name     string
			v, other Vec2i
			expected Vec2i
		}{
			{"simple diff", Vec2i{5, 5}, Vec2i{2, 2}, Vec2i{3, 3}},
			{"wrap around right", Vec2i{2, 2}, Vec2i{8, 8}, Vec2i{4, 4}},
			{"wrap around left", Vec2i{8, 8}, Vec2i{2, 2}, Vec2i{-4, -4}},
			{"wrap around top", Vec2i{5, 2}, Vec2i{5, 8}, Vec2i{0, 4}},
			{"wrap around bottom", Vec2i{5, 8}, Vec2i{5, 2}, Vec2i{0, -4}},
			{"large diff positive", Vec2i{25, 25}, Vec2i{2, 2}, Vec2i{3, 3}},
			{"large diff negative", Vec2i{2, 2}, Vec2i{25, 25}, Vec2i{-3, -3}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := tt.v.DiffP(tt.other)
				if got != tt.expected {
					t.Errorf("%v.DiffP(%v) = %v; expected %v", tt.v, tt.other, got, tt.expected)
				}
			})
		}
	})
}
