package game

import (
	"bytes"
	"fmt"
	"testing"

	"encoding/gob"

	"github.com/vmihailenco/msgpack/v5"
)

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
			{
				Blabla:    "ListItem1",
				Bliblib:   7,
				Lol:       []float64{1.618},
				Trollolol: map[string]int{"three": 3},
			},
			{
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

	fmt.Println(tst)
	fmt.Println(tst.Sub)
	fmt.Println(tst.Kasdf[0])

	//t.Fail()
}

func TestGOB(t *testing.T) {
	original := &TestStruct2{
		Blabla: "Hello, World!",
		Sub: &TestStruct1{
			Blabla:    "SubStruct",
			Bliblib:   42,
			Lol:       []float64{3.14, 2.718},
			Trollolol: map[string]int{"one": 1, "two": 2},
		},
		Kasdf: []*TestStruct1{
			{
				Blabla:    "ListItem1",
				Bliblib:   7,
				Lol:       []float64{1.618},
				Trollolol: map[string]int{"three": 3},
			},
			{
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
	fmt.Println(tst)
	fmt.Println(tst.Sub)
	fmt.Println(tst.Kasdf[0])

	//t.Fail()
}
