package game

import (
	"testing"
)

func TestPlayerSnake_GetOwner(t *testing.T) {
	// Setup a minimal PlayerSnake
	base := &BaseSnake{}
	ps := &PlayerSnake{BaseSnake: base, ID: 1}

	// Verify GetOwner returns the PlayerSnake, not BaseSnake
	owner := ps.GetOwner()
	if _, ok := owner.(*PlayerSnake); !ok {
		t.Errorf("GetOwner() returned type %T, expected *PlayerSnake", owner)
	}
}
