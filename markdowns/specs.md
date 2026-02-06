# SnakePro Game Overview
SnakePro is a local split-screen evolution of classic snake. Multiple players guide their snakes across a shared grid, competing for survival, items, and clever positioning while avoiding collisions with walls, other snakes, and hostile entities.

## Design Pillars
- **Deterministic core**: pure game logic drives authoritative state, enabling replays and tooling while keeping rendering separate.
- **Data-driven tuning**: gameplay constants live in config files so variants, mods, and balance patches require no recompilation.
- **Extensibility**: new items, entities, and map behaviors plug into shared interfaces without refactoring existing systems.


## Resource system

### User data
- Configuration files are stored in a platform-appropriate user data directory `/usr/share/snakeprogo/config/`.
- Config files are in TOML format and define gameplay constants, and user preferences/stats (e.g. keybindings, name, stats).
- Since SnakePro is a local multiplayer game, each user has a separate config file named after their username (e.g. `alice.toml`), containing their personal settings.
- The main configuration file is `config.toml`, which contains global settings like default map & assets, graphics options, and audio levels.
- The gamplay constants are stored in `gameplay.toml`, which can be modified to tweak game balance and mechanics.
- The folder structure is as follows:
  ```
  /usr/share/snakeprogo/config/
  ├── config.toml
  ├── gameplay.toml
  └── userconfig/
	  ├── alice.toml
	  └── bob.toml
  ```
- Config files are loaded at startup (and created with defaults if missing)
- No GUI for editing config files is provided; users must edit them manually with a text editor.


### Asset Management
- A resource manager loads all assets (images, sounds, maps) at startup and provides global access via string keys.
- Maps are stored as JSON files defining tile layouts, spawn points, and hazard zones.
- Folders with suffix `_se` are loaded as a single Soundeffect resource that randomly picks one audio clip on play.
- By default the game will look for assets in `/usr/share/snakeprogo/res/`.
- In the main configuration file, the user can alternatively specify a custom assets directory. The game will then treat that as the root for all asset loading instead of the default "res" folder.


## Game Logic Layer
All gameplay code must operate on the authoritative simulation state. Key subsystems:

### Game Session
- Runs the fixed-step update loop covering movement, physics-style collision checks, input dispatch, map updates, and statistics tracking.
- Owns current game state, active map instance, and player/channel bindings.

### Game State
- Holds precise player snapshots (position history, velocity, status effects), entity data, and map tiles.
- Renderer, replay exporter, and analytics read exclusively from this structure.

### Statistics
- Aggregates per-player metrics (kills, deaths, item usage, distance traveled) for post-match screens and meta progression.

### Input Abstraction
- Accepts arbitrary devices but normalizes to: `turn-left`, `turn-right`, `up`, `down`, `left`, `right`, and `use-item`.
- Supports both relative (turning) and absolute cardinals so controllers and keyboards feel native.

### Map System
- Each map is a scene-aware game object updated every tick.
- Responsibilities include collision shapes, bounds logic (wrapping, hard walls, hybrid regions), spawn points, and optional hazards.

### Consumables and Entities
- **Apples**: spawn randomly, extend snakes by increasing the "fett" counter that controls tail removal.
- **Items**: collectible abilities that can be triggered asynchronously; most spawn one or more entities.
- **Entities**: autonomous actors (bullets, bots, AoE fields) updated independently yet interacting with snakes and terrain.


## Visual Layer
Using Ebiten render everything directly from the authoritative game state. No visual component may mutate gameplay data, enabling deterministic replays and remote spectators. The game is rendered to a single window with the entire map on screen.

## Core Loop
1. Process buffered input: change heading, trigger items if available.
2. Advance snakes: move heads; if `fett > 0`, keep the tail and decrement the counter, otherwise pop tail segments.
3. Resolve collisions: snake vs. map, snake vs. snake, entity interactions.
4. Update map and entities (bombs, bullets, fart clouds, etc.).
5. Apply consumable effects: apple nutrition increases `fett`, item pickups assign inventory and fire `on_collect` hooks.
6. Evaluate victory/timeout rules (e.g., end after 30 ticks with <2 living players).


## Death and Respawn Rules
- Colliding with any non-head snake segment or impassable tile kills the snake.
- Configurable modes:
  - **Elimination**: player remains out until the session ends.
  - **Ghost Mode**: player respawns as a ghost snake that cannot collect items/apples or kill, but can rot consumables it touches.
- Ghost snakes convert back to living snakes via the Revive item; if they die while holding it, the effect applies immediately on respawn.


## Configuration Schema
All tunables load from the primary config before the match begins. Example keys:
- `startSnakeLength`
- `snakeSpeed`
- `mapPath`
- `appleCount`
- `appleNutrition`
- `appleRotTime`
- `ghostAppleDamage`
- `itemCount`

## Item Reference

### Shooting
Fires a bullet along the facing direction that carves snake bodies. Hits on the body prune all segments after the impact point; headshots kill outright. Snakes that shrink below `snakeSurvivalLength` die.
- `bulletSpeed`
- `bulletRange`
- `snakeSurvivalLength`

### Speed Boost
Temporarily increases the caster's movement speed.
- `speedMultiplier`
- `duration`

### Bot Strike
Spawns an autonomous kamikaze bot using simple A*/directional pathfinding to cut off the nearest opponent.
- `botSpeed`
- `botLength`
- `botDuration`

### Fart Field
Creates a lingering square zone at the snake's tail that drains length from enemies passing through. While `fett` remains, only the counter drops; otherwise tail pieces are removed. Snakes below length 2 that would lose another segment die instantly.
- `fartDuration`
- `fartSize`
- `fartDamagePerSecond`

### Revive
Returns ghost snakes to life or lets living snakes instantly respawn at their spawn point, clearing status effects and resetting length. Auto-triggers if the holder dies.