# SnakeProGo

A feature-packed Snake game implementation in Go using the [Ebitengine](https://ebitengine.org/) 2D game library.

## Installation
download the installer script and run it.

## Features

- **Classic Gameplay**: Navigate the snake, eat apples, and grow.
- **Advanced Mechanics**:
  - **Items & Power-ups**: Speed boosts, Revive/Rewind capability, and Shooting mechanics.
  - **Combat**: Shoot bullets to clear obstacles or threats.
  - **Status Effects**: Various effects that impact gameplay.
  - **Replay System**: Save and watch replays of your best runs.
  - **LLM Integration**: Uses AI to generate creative and unique filenames for your replay saves.
- **Configuration**: Highly customizable gameplay via TOML configuration.
- **Input**: Support for Keyboard and Gamepad controllers with custom mapping.

## Requirements

- [Ollama](https://ollama.com/) - Required for generating fancy replay names via LLM.
- [FFmpeg](https://ffmpeg.org/) - Required for saving replays to video files.

## Build requirements
- Go 1.24 or higher
- C Compiler (gcc/clang) - required by Ebitengine for low-level graphics API bindings.

## License

See [LICENSE](LICENSE) file.
