package game

import (
	"bytes"
	"fmt"
	"testing"

	"encoding/gob"

	"github.com/vmihailenco/msgpack/v5"
)

func TestGameStateMarshalling(t *testing.T) {
	state := &GameState{
		Tick: 12345,
		Map: &MapData{
			Collider: &CollisionMap{
				UseBounds: true,
				Width:     20,
				Height:    15,
				P0:        Vec2i{X: 0, Y: 0},
				Occupied:  [][]bool{},
			},
		},
		Players: map[int]*PlayerSnake{
			1: NewPlayerSnake(
				NewBaseSnake(Vec2i{0, 4}, DirLeft, 5),
				52,
				&PlayerConfig{
					Name: "TestPlayer",
				},
			),
			2: NewPlayerSnake(
				NewBaseSnake(Vec2i{0, 4}, DirLeft, 5),
				52,
				&PlayerConfig{
					Name: "TestPlayer",
				},
			),
		},
		Apples: []*Apple{
			NewApple(1, Vec2i{X: 7, Y: 8}),
		},
		Items: []*Item{
			NewItem(1, Vec2i{X: 9, Y: 10}, ItemRevive),
		},
		Entities: []Entity{
			&EntityBase{
				ID:   1,
				Type: EntityBullet,
				Collider: &CollisionTiles{
					Tiles: []Vec2i{{X: 11, Y: 12}},
				},
				OwnerID:  1,
				LifeTime: 5,
			},
		},
	}

	b, err := state.MarshalAllObjects()
	if err != nil {
		t.Fatal("Encoding error:", err)
	}

	state.Tick = 18763
	state.Players[1].Facing = DirDown
	state.Players[1].StatusEffects = []*StatusEffect{NewDeadStatusEffect()}
	state.Events = []*GameEvent{NewSoundEvent("TestSound")}

	deltab, err := state.MarshalMutableObjects()
	if err != nil {
		t.Fatal("Delta encoding error:", err)
	}

	var decodedState GameState
	err = decodedState.UnmarshalAllObjects(b)
	if err != nil {
		t.Fatal("Decoding error:", err)
	}

	fmt.Printf("Original state: %+v\n", state)
	fmt.Printf("tick: %d, player 1 facing: %v, status effects: %v, events: %v\n", state.Tick, state.Players[1].Facing, state.Players[1].StatusEffects, state.Events)
	fmt.Printf("Decoded state: %+v\n", decodedState)
	fmt.Printf("tick: %d, player 1 facing: %v, status effects: %v, events: %v\n", decodedState.Tick, decodedState.Players[1].Facing, decodedState.Players[1].StatusEffects, decodedState.Events)

	err = decodedState.UnmarshalMutableObjects(deltab)
	if err != nil {
		t.Fatal("Delta decoding error:", err)
	}

	fmt.Printf("After delta decode: %+v\n", decodedState)
	fmt.Printf("tick: %d, player 1 facing: %v, status effects: %v, events: %v\n", decodedState.Tick, decodedState.Players[1].Facing, decodedState.Players[1].StatusEffects, decodedState.Events)

	t.Fail()
}

type TestStruct1 struct {
	Blabla    string `msgpack:"-"`
	Bliblib   int
	Lol       []float64
	Trollolol map[string]int
}

type TestStruct2 struct {
	Blabla string
	Sub    *TestStruct1
	Kasdf  []*TestStruct1
}

func TestMarshalling(t *testing.T) {
	original := &TestStruct2{
		Blabla: "Hello, World!",
		Sub: &TestStruct1{
			Blabla:    "SubStruct",
			Bliblib:   42,
			Lol:       []float64{3.14, 2.718},
			Trollolol: map[string]int{"one": 1, "two": 2},
		},
		Kasdf: []*TestStruct1{
			&TestStruct1{
				Blabla:    "ListItem1",
				Bliblib:   7,
				Lol:       []float64{1.618},
				Trollolol: map[string]int{"three": 3},
			},
			&TestStruct1{
				Blabla:    "ListItem2",
				Bliblib:   13,
				Lol:       []float64{0.577},
				Trollolol: map[string]int{"four": 4},
			},
		},
	}
	b, err := msgpack.Marshal(original)
	if err != nil {
		panic(err)
	}
	var tst TestStruct2
	err = msgpack.Unmarshal(b, &tst)
	if err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Println(tst)
	fmt.Println()
	fmt.Println(tst.Sub)
	fmt.Println()
	fmt.Println(tst.Kasdf[0])
	fmt.Println()

	//t.Fail()
}

func TestGOB(t *testing.T) {
	gob.Register(&TestStruct1{})
	gob.Register(&TestStruct2{})
	original := &TestStruct2{
		Blabla: "Hello, World!",
		Sub: &TestStruct1{
			Blabla:    "SubStruct",
			Bliblib:   42,
			Lol:       []float64{3.14, 2.718},
			Trollolol: map[string]int{"one": 1, "two": 2},
		},
		Kasdf: []*TestStruct1{
			&TestStruct1{
				Blabla:    "ListItem1",
				Bliblib:   7,
				Lol:       []float64{1.618},
				Trollolol: map[string]int{"three": 3},
			},
			&TestStruct1{
				Blabla:    "ListItem2",
				Bliblib:   13,
				Lol:       []float64{0.577},
				Trollolol: map[string]int{"four": 4},
			},
		},
	}
	var network bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&network) // Will write to network.
	dec := gob.NewDecoder(&network) // Will read from network.

	// Encode (send) some values.
	err := enc.Encode(original)
	if err != nil {
		t.Fatal("encode error:", err)
	}

	// Decode (receive) and print the values.
	var tst TestStruct2
	err = dec.Decode(&tst)
	if err != nil {
		t.Fatal("decode error:", err)
	}

	fmt.Println()
	fmt.Println(tst)
	fmt.Println()
	fmt.Println(tst.Sub)
	fmt.Println()
	fmt.Println(tst.Kasdf[0])
	fmt.Println()

	//t.Fail()
}
