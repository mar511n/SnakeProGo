package main

import (
	"SnakeProGo/game"
	"fmt"
	"math/rand/v2"
)

func main() {
	game.InitLLM()
	hash := rand.Uint32()
	adj1, evt, adj2, par := game.RandomizeFilenamePartsSimple(hash)
	fname1 := fmt.Sprintf("%s %s %s %s", adj1, evt, adj2, par)
	fmt.Println("Generated filename:", fname1)
	fname2 := game.GenerateFilenameForReplay(fname1)
	fmt.Println("LLM generated filename:", fname2)

	for i := 0; i < 0; i++ {
		hash := rand.Uint32()
		adj1, evt, adj2, par := game.RandomizeFilenamePartsSimple(hash)
		fname1 := fmt.Sprintf("%s %s %s %s", adj1, evt, adj2, par)
		fmt.Println(fname1)
	}
}
