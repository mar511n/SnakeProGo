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

## Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd SnakeProGo
   ```

2. Download dependencies:
   ```bash
   go mod tidy
   ```

## Running the Game

To start the game from the source:

```bash
go run src/cmd/main.go
```

## Configuration

The game relies on a configuration system defined in `src/game/config.go`. It supports loading settings from a TOML file, allowing you to tweak:

- **Display**: Screen width, height, fullscreen mode, VSync.
- **Audio**: Master, Music, and SFX volume levels.
- **Gameplay**:
  - Snake speed and starting length.
  - Apple count and nutrition value.
  - Item spawn rates (Speed, Revive, Shooting).
  - Invincibility and Rewind mechanics.

## LLM Integration

The project includes an LLM-powered utility for generating descriptive filenames. You can verify this component specifically by running:

```bash
go run src/cmd/llm.go
```

## Project Structure

- `src/cmd/`: Application entry points (`main.go`, `llm.go`).
- `src/game/`: Core game logic, entities (Snake, Apple, Bullet), and systems (Renderer, Input, GameState).
- `src/llm/`: Logic for Large Language Model integration.
- `src/assets/`: (Expected) Directory for game assets like images and sounds.

## License

See [LICENSE](LICENSE) file.
