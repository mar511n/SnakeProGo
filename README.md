# SnakeProGo

A feature-packed Snake game implementation in Go using the [Ebitengine](https://ebitengine.org/) 2D game library.

## Installation
Download the latest installer for your platform:
- [Linux](https://github.com/mar511n/SnakeProGo/releases/latest/download/SnakeProInstaller-linux-amd64.zip)
- [macOS](https://github.com/mar511n/SnakeProGo/releases/latest/download/SnakeProInstaller-darwin-arm64.zip)
- [Windows](https://github.com/mar511n/SnakeProGo/releases/latest/download/SnakeProInstaller-windows-amd64.zip)

Extract and run the installer to set up the game on your system.

## Features

- **Classic Gameplay**: Navigate the snake, eat apples, and grow.
- **Advanced Mechanics**:
  - **Items & Power-ups**: Speed boosts, Revive/Rewind capability, and Shooting mechanics.
  - **Combat**: Shoot bullets to clear obstacles or threats.
  - **Status Effects**: Various effects that impact gameplay.
  - **Replay System**: Save and watch replays of your best runs.
- **Configuration**: Highly customizable gameplay via TOML configuration.
- **Input**: Support for Keyboard and Gamepad controllers with custom mapping.

## Requirements

### Build requirements
- Go 1.24 or higher
- C Compiler (gcc/clang) - required by Ebitengine for low-level graphics API bindings.

### Optional dependencies
- [FFmpeg](https://ffmpeg.org/) - Required for saving replays to video files.
- [Ollama](https://ollama.com/) - Required for generating fancy replay names via LLM.

## License

See [LICENSE](LICENSE) file.
