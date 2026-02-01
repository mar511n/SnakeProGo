# Game Backend Implementation Guide

## TODO

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

// Standard Directions
var (
    DirUp    = Point{0, -1}
    DirDown  = Point{0, 1}
    DirLeft  = Point{-1, 0}
    DirRight = Point{1, 0}
)

// basic operations on Point
func (p Point) Add(other Point) Point 
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
    Points []Point // Slice of values for cache locality
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
    Width, Height int
    Occupied      [][]bool
}

func (c *CollisionMap) Contains(p Point) bool {
    if p.X < 0 || p.X >= c.Width || p.Y < 0 || p.Y >= c.Height {
        return true // Treat out of bounds as collision (or handle wrapping)
    }
    return c.Occupied[p.X][p.Y]
}

func (c *CollisionMap) IsColliding(other CollisionObject) bool {
    switch o := other.(type) {
    case *CollisionTiles:
        return o.IsColliding(c) // Delegate
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
    EntityBullet EntityType = iota
    EntityBot
    EntityApple
    EntityItem
    EntityFartCloud
)

// Entity represents any dynamic object in the world (except PlayerSnakes).
// Collectibles (Items & Apples), bots, bullets and farts inherit from this.
// Implements Collidable and Updatable interfaces.
type Entity struct {
    ID             uint64
    Type           EntityType
    Collider       *CollisionTiles  // The entity's physical presence
    OwnerID        int              // Player who spawned it (or 0 for world)
    LifeTime       int              // Ticks remaining (-1 for infinite)
    Behavior       Updatable        // specific logic (e.g., bot AI, missile tracking)
}

// used by the item struct to identify itself
type ItemType int

const (
    ItemShooting ItemType = iota
    ItemSpeed
    ItemBot
    ItemFart
    ItemRevive
)
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
    Invulnerable    int             // Ticks remaining
}

// PlayerSnake represents a player-controlled snake.
type PlayerSnake struct {
    *BaseSnake
    ID              int
    Config          *PlayerConfig // Reference to existing PlayerConfig struct (name, keys, stats)
}
```

### 5. Map System (`map.go`)

The map defines the playable area, obstacles, and spawn points.

**Structs:**

```go
type TileType int

const (
    TileEmpty TileType = iota
    TileWall
    TileHazard
    TileSpawn
)

// implements Collidable
type MapData struct {
    Width, Height int
    Tiles         [][]TileType
    Collider      *CollisionMap    // Optimised collision map
    SpawnPoints   []Point          // Slice of values
}
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
    Apples      []*Entity            // Collectible apples
    Items       []*Entity            // Collectible items
    Entities    []*Entity            // Dynamic entities (Bullets, Farts, Bots)
    Events      []*GameEvent
}

type GameSession struct {
    Config            *GameplayConfig            // Reference to global GameplayConfig
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
2.  **Update PlayerSnakes**: Call the Update method on each PlayerSnake.
3.  **Entity Updates**:
    - Iterate `State.Apples`, `State.Items`, and `State.Entities`: Call `e.Update(State)`.
    - Handle `LifeTime` expiration.
4.  **Collision loop**:
    - For each `Collidable` (Snakes, Entities, Map Tiles), check for collisions between objects that share at minimum one collision layer.
    - Call `OnCollision` on both objects when a collision is detected.
5.  **Spawning**:
    - Check `len(State.Apples)` and `len(State.Items)`. Spawn new Apple/Item if below target counts.
6.  **Win Condition**: Check living snakes.
7.  **History**: Serialize current `GameState` and append to `History`.

## Next Steps

