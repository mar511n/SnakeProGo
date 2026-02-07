# Game Backend Implementation Guide

This document outlines the architecture and data structures required to implement the authoritative game backend for SnakeProGo. The backend is responsible for maintaining the simulation state, processing input, resolving collisions, and managing the core game loop.

## Core Architecture

The backend will be structured around a deterministic `GameSession` that updates the `GameState` at a fixed time step (TPS).

### 1. Fundamental Types (`geometry.go`)

Basic geometric types and interfaces used throughout the backend.

**Structs:**

```go
// Point represents a grid coordinate or tile.
type Point struct {
    X, Y int
}

var (
    DirUp    = Point{0, -1}
    DirDown  = Point{0, 1}
    DirLeft  = Point{-1, 0}
    DirRight = Point{1, 0}
)

func (p Point) Add(other Point) Point 
func (p Point) Sub(other Point) Point
func (p Point) Eq(other Point) bool

// Vector represents spatial velocity or direction or displacement.
type Vector struct {
    X, Y float64
}

func (v Vector) ToPoint() Point
func (v Vector) Add(other Vector) Vector
func (v Vector) Mul(scalar float64) Vector

// Updatable interface for objects that need per-tick updates
type Updatable interface {
    Update(state *GameState)
}

// GameEvent used for broadcasting events to the frontend
type GameEvent struct {
    ID      uint64
    Type    string
    Payload interface{}
}
```

### 2. Collision System (`collision.go`)

The collision system defines how objects interact.

**Structs:**

```go
// objects can be on any of 16 collision layers
type CollisionMask uint16

const (
    LayerNone   CollisionMask = 0
    LayerWall   CollisionMask = 1 << 0
    LayerSnake  CollisionMask = 1 << 1
    LayerApple  CollisionMask = 1 << 2
    LayerItem   CollisionMask = 1 << 3
    LayerHazard CollisionMask = 1 << 4
    LayerEntity CollisionMask = 1 << 5
    LayerProjectile CollisionMask = 1 << 6
)

type Collidable interface {
    OnCollision(other Collidable, state *GameState) // Handle collision with another object
    GetCollisionMask() CollisionMask                // The collision layers this object belongs to
    GetCollider() CollisionObject                   // Returns the geometric shape of the object
}

type CollisionObject interface {
    IsColliding(other CollisionObject) bool
}

// CollisionTiles implements CollisionObject for sparse objects (Entities, Snakes)
type CollisionTiles struct {
    Points []Point
}

func (c *CollisionTiles) IsColliding(other CollisionObject) bool {
    switch o := other.(type) {
    case *CollisionTiles:
        // Check for point overlaps O(N*M)
        for _, p1 := range c.Points {
            for _, p2 := range o.Points { 
                if p1.Eq(p2) { return true }
            }
        }
    case *CollisionMap:
        // Check if any point is in map bounds/walls
        for _, p := range c.Points {
            if o.Contains(p) { return true }
        }
    }
    return false
}

// CollisionMap implements CollisionObject for static map geometry
type CollisionMap struct {
    UseBounds bool
    P0 Point
    Width,Height int
    Occupied      [][]bool
}

func (c *CollisionMap) Contains(p Point) bool {
    pr := p.Sub(c.P0)
    if pr.X < 0 || pr.Y < 0 || pr.X >= c.Width || pr.Y >= c.Height {
        return c.UseBounds
    }
    return c.Occupied[pr.X][pr.Y]
}

func (c *CollisionMap) IsColliding(other CollisionObject) bool {
    switch o := other.(type) {
    case *CollisionTiles:
        return o.IsColliding(c)
    }
    return false
}
```

### 3. Entity System (`entity.go`)

The `Entity` struct is the generic container for all dynamic objects in the game world, including projectiles, bots, apples, and items.

**Structs:**

```go
type EntityType int

const (
    EntityApple EntityType = iota
    EntityItem
    EntityBullet
    EntityBot
    EntityFartCloud
)

// Entity interface ensures objects satisfy collision and update protocols.
type Entity interface {
    Collidable
    Updatable
}

// EntityBase represents any dynamic object in the world.
// Entities (Apples, Items, Bullets, Bots) embed this.
type EntityBase struct {
    ID             uint64
    Type           EntityType
    Collider       *CollisionTiles  // The entity's physical presence
    OwnerID        int              // Player who spawned it (or -1 for world)
    LifeTime       int              // Ticks remaining (-1 for infinite)
}

func (e *EntityBase) Update(state *GameState) {
    if e.LifeTime > 0 {
        e.LifeTime--
    }
}

// Apple embeds EntityBase and adds specific logic if needed (e.g. what happens on collision).
type Apple struct {
    *EntityBase
}

// used by the item struct to identify itself
type ItemType int

const (
    ItemNone ItemType = iota
    ItemShooting
    ItemSpeed
    ItemBot
    ItemFart
    ItemRevive
)

// ItemHandler defines the effect of using an item. 
// It returns true if the item was successfully used (and should be consumed).
type ItemHandler func(user *PlayerSnake, state *GameState) bool

// Global registry of item behaviors, populated at startup.
var ItemRegistry = map[ItemType]ItemHandler{}

type Item struct {
    *EntityBase
    ItemType       ItemType
}
```

### 4. Snake Actors (`snake.go`)

Snakes are the primary actors. We separate the core snake logic from player-specific data.

**Structs:**

```go
// BaseSnake contains the physical properties and status of a snake entity.
// Implements Collidable and Updatable.
// Update method handles movement, growth, and status effects updates.
type BaseSnake struct {
    Body            *CollisionTiles // Head is at index 0 of the Points slice
    Facing          *Point          // Current movement direction
    NextFacing      *Point          // Buffered input direction
    Fett            int             // "Fett" counter for growth buffer
    StatusEffects   []Updatable     // Active status effects
    IsGhost         bool            // Can pass through walls
    IsDead          bool
    Invulnerable    int             // Ticks remaining (0 = no)
}

// PlayerSnake represents a player-controlled snake.
type PlayerSnake struct {
    *BaseSnake
    ID              int
    Config          *PlayerConfig // Reference to existing PlayerConfig struct (name, keys, stats)
    HeldItem        ItemType      // Currently held item (ItemNone if empty)
}
```

### 5. Map System (`map.go`)

The map defines the playable area, obstacles, and spawn points.

**Structs:**

```go
// enables differently rendered tiles and collision properties
type Tile struct {
    Name    string
    IsWall  bool
    IsSpawn bool
}

// implements Collidable
type MapData struct {
    Tiles         [][]Tile
    Collider      *CollisionMap    // Optimised collision map (contains width, height)
    SpawnPoints   []Point          // Slice of values
}

// BuildCache populates Collider and SpawnPoints from the Tiles grid
func (m *MapData) BuildCache()
```

### 6. Input System (`input.go`)

**Structs:**

```go
type PlayerAction int

const (
    ActionNone PlayerAction = iota
    ActionUp
    ActionDown
    ActionLeft
    ActionRight
    ActionTurnLeft
    ActionTurnRight
)

// InputFrame captures inputs for a specific tick.
type InputFrame struct {
    Tick         uint64
    Directions   map[int]PlayerAction // Keyed by player ID
    ItemsUsed    map[int]bool         // Keyed by player ID, true indicates item usage
}

// Process reads current hardware inputs and populates Directions and ItemsUsed based on player keymaps.
func (i *InputFrame) Process(playerConfigs map[int]*PlayerConfig)
```

### 7. Game State & Session (`state.go`, `session.go`)

**Structs:**

```go
// GameState holds the complete state of the simulation at a specific tick.
type GameState struct {
    Tick        uint64
    Map         *MapData
    Snakes      map[int]*PlayerSnake // Keyed by player ID
    Apples      []Entity             // Collectible apples
    Items       []Entity             // Collectible items
    Entities    []Entity             // Dynamic entities (Bullets, Farts, Bots)
    Events      []*GameEvent
}

type GameSession struct {
    State             *GameState
    RegisteredPlayers map[int]*PlayerConfig      // Key: ID, Value: Reference to PlayerConfig
    Input             *InputFrame                // Current frame inputs
    RandSeed          int64
    Winner            int
    History           [][]byte                   // List of serialized GameStates for replay
}
```

### 8. Configuration Extensions

To bridge the gap between user-friendly configuration units (floating-point speeds in tiles/sec, durations in seconds) and the integer-based fixed-step simulation, configuration structs will implement getter methods.

**Global Access:**
The backend logic accesses the global `GPConfig` variable (defined in `config.go`) directly to retrieve gameplay constants during updates and initialization.

**Helpers:**

```go
// Calculate ticks per move based on speed (tiles/sec) and TPS
func (c *GameplayConfig) GetTicksPerMove(speed float64, tps int) int

// Convert duration in seconds to ticks
func (c *GameplayConfig) GetDurationTicks(seconds float64, tps int) int
```

## Process Flow

### 1. Initialization
- Load Map.
- Initialize `GameState`.
- Spawn initial Apples/Items as Entities.
- Place Snakes.

### 2. The Update Loop (`GameSession.Update()`)
1.  **Read Input**: `Session.Input.Process(Session.RegisteredPlayers)` to capture current actions.
2.  **Item Usage**:
    - For each player `ID` with `Input.ItemsUsed[ID] == true`:
        - If `Snakes[ID].HeldItem != ItemNone`, look up logic in `ItemRegistry`.
        - Execute logic. If returns true, set `HeldItem = ItemNone`.
3.  **Update PlayerSnakes**: Call the Update method on each PlayerSnake.
4.  **Entity Updates**:
    - Iterate `State.Apples`, `State.Items`, and `State.Entities`: Call `e.Update(State)`.
    - Handle `LifeTime` expiration.
4.  **Collision loop**:
    - For each `Collidable` (Snakes, Apples, Items, Entities, Map Tiles), check for collisions between objects that share at minimum one collision layer.
        - proceed in the following order:
            1. Snakes vs Map
            2. Snakes vs Snakes
            3. Snakes vs Apples/Items
            4. Entities vs Map
            5. Entities vs Entities
            6. Entities vs Apples/Items
            7. Entities vs Snakes
    - Call `OnCollision` on both objects when a collision is detected.
5.  **Spawning**:
    - Check `len(State.Apples)` and `len(State.Items)`. Spawn new Apple/Item if below target counts.
6.  **Win Condition**: Check living snakes.
7.  **History**: Serialize current `GameState` and append to `History`.

## Next Steps

