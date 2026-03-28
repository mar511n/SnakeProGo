package game

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/prompts"
)

var (
	LLM                     *ollama.LLM
	LLM_context             context.Context
	LLM_gen_filename_prompt prompts.PromptTemplate
)

func InitLLM() {
	if !GConfig.UseLLM {
		LogInfo("LLM usage is disabled in config.")
		return
	} else {
		LogInfo("Initializing ollama model qwen3:0.6b...")
	}
	llm, err := ollama.New(ollama.WithModel("qwen3:0.6b"))
	if err != nil {
		LogError("Failed to initialize LLM: %v", err)
	}
	LLM = llm
	LLM_context = context.Background()
	LLM_gen_filename_prompt = prompts.NewPromptTemplate(
		`You are a creative writer. Turn these random words into an Epic Movie Title for a snake game.
Make it dramatic. You can reorder or slightly change words to make it flow better.
The filename must end with .mp4 and use underscores instead of spaces.

Examples:
Input: spectral attack of golden anacondas
Output: Attack_of_the_Golden_Anacondas.mp4

Input: incredible feast of stygian serpents
Output: The_Incredible_Feast_of_Stygian_Serpents.mp4

Input: tiny hunt of ethereal cobras
Output: Hunt_of_the_Ethereal_Cobras.mp4

Input: obsidian clash of insane worms
Output: Clash_of_the_Insane_Obsidian_Worms.mp4

Input: vengeful rise of hallowed lizards
Output: The_Vengeful_Rise_of_Hallowed_Lizards.mp4

Current Task:
Input: {{.filename}}
Output:`,
		[]string{"filename"},
	)
}

func GenerateFilenameForReplay(filename string) string {
	simple_filename := strings.ReplaceAll(filename, " ", "_")
	if !strings.HasSuffix(simple_filename, ".mp4") {
		simple_filename += ".mp4"
	}
	if LLM == nil {
		if GConfig.UseLLM {
			LogError("LLM not initialized. Make sure ollama is installed with the model qwen3:0.6b")
		}
		return simple_filename
	}

	promptvalue, err := LLM_gen_filename_prompt.FormatPrompt(map[string]any{
		"filename": filename,
	})
	if err != nil {
		LogError("Failed to format prompt: %v", err)
		return simple_filename
	}
	// We use a low temperature to make the model deterministic.
	// We also don't process the reasoning part, so we just take the last line.
	response, err := LLM.Call(LLM_context, promptvalue.String(), llms.WithTemperature(0.45), llms.WithStopWords([]string{"Input:"}), llms.WithMaxTokens(1000), llms.WithThinkingMode(llms.ThinkingModeHigh))
	if err != nil {
		LogError("Failed to call LLM: %v", err)
		return simple_filename
	}
	newname := simple_filename
	lines := strings.Split(response, "\n")
	if len(lines) == 0 {
		LogError("LLM returned empty response.")
		return simple_filename
	}
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			newname = lines[i]
			break
		}
	}
	newname = lines[len(lines)-1]
	newname = strings.ReplaceAll(newname, " ", "_")
	newname = strings.ReplaceAll(newname, "-", "_")
	newname = strings.ToLower(newname)
	if !strings.HasSuffix(newname, ".mp4") {
		newname += ".mp4"
	}
	if strings.ContainsAny(newname, `\/:*?"<>|`) || len(newname) < 6 || len(newname) > 100 {
		LogError("LLM generated filename is invalid: %s", newname)
		return simple_filename
	}
	LogInfo("LLM response for filename generation: %s", newname)
	return newname
}

var adjectives = []string{
	"epic", "incredible", "slimy", "wild", "furious", "sneaky", "giant", "tiny", "hungry", "angry",
	"green", "dangerous", "fast", "ancient", "mystic", "dark", "golden", "silent", "brave", "cruel",
	"crazy", "insane", "bloody", "chaotic", "peaceful", "intense", "savage", "fearsome", "magical", "venomous",
	"eldritch", "stygian", "inexorable", "voracious", "pernicious", "insidious", "serried", "labyrinthine", "arcane",
	"primordial", "spectral", "phantasmal", "cryptic", "ominous", "malevolent", "treacherous", "vengeful", "abyssal",
	"lurid", "hallowed", "unseen", "shadowy", "ethereal", "tempestuous", "perfidious", "obsidian", "emerald", "crimson",
}

var events = []string{
	"battle of", "clash of", "dance of", "escape of", "hunt of", "journey of", "legend of", "mystery of",
	"rise of", "fall of", "attack of", "revenge of", "duel of", "invasion of", "panic of", "feast of",
	"chase of", "race of", "doom of", "awakening of",
}

var participants = []string{
	"anacondas", "pythons", "vipers", "cobras", "worms", "noodles", "serpents", "reptiles", "dragons", "lizards",
	"boas", "rattlers", "mambas", "asps", "constrictors", "hydras", "basilisks", "sidewinders", "sneks", "predators",
}

func RandomizeFilenamePartsSimple(hash uint32) (adj1, evt, adj2, par string) {
	lAdj := uint32(len(adjectives))
	lEvt := uint32(len(events))
	lPar := uint32(len(participants))

	idx1 := (hash) % lAdj
	idx2 := (hash / lAdj) % lEvt
	idx3 := (hash / (lAdj * lEvt)) % lAdj
	idx4 := (hash / (lAdj * lEvt * lAdj)) % lPar
	return adjectives[idx1], events[idx2], adjectives[idx3], participants[idx4]
}

func RandomizeFilenameSimple(hash uint32) string {
	// Pick words based on the hash
	// Use the hash to seed a simple generator

	adj1, evt, adj2, par := RandomizeFilenamePartsSimple(hash)
	evt = strings.ReplaceAll(evt, " ", "_")
	return fmt.Sprintf("%s_%s_%s_%s.mp4", adj1, evt, adj2, par)
}
