# SnakeProGo

A feature-packed Snake game implementation in Go using the [Ebitengine](https://ebitengine.org/) 2D game library.

## Features

- **Classic Gameplay**: Navigate the snake, eat apples, and grow.
- **Advanced Mechanics**:
  - **Items & Power-ups**: Speed boosts, Revive/Rewind capability, and Shooting mechanics.
  - **Combat**: Shoot bullets to clear obstacles or threats.
  - **Status Effects**: Various effects that impact gameplay.
  - **Replay System**: Save and watch replays of your best runs.
  - **LLM Integration**: Uses AI to generate creative and unique filenames for your replay saves.
- **Configuration**: Highly customizable gameplay via TOML configuration (`game/config.go`).
- **Input**: Support for Keyboard and Gamepad controllers with custom mapping.

## Requirements

- Go 1.24 or higher
- C Compiler (gcc/clang) - required by Ebitengine for low-level graphics API bindings.
- [Ollama](https://ollama.com/) - Required for generating fancy replay names via LLM.
- [FFmpeg](https://ffmpeg.org/) - Required for saving replays to video files.


## Project Structure

- `src/cmd/`: Application entry points (`main.go`, `llm.go`).
- `src/game/`: Core game logic, entities (Snake, Apple, Bullet), and systems (Renderer, Input, GameState).
- `src/llm/`: Logic for Large Language Model integration.
- `src/assets/`: (Expected) Directory for game assets like images and sounds.

## License

See [LICENSE](LICENSE) file.
