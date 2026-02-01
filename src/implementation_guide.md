# Implementation guide

- start with a simple CLI like main menu (using ebitenutil.DebugPrint and inpututil.IsKeyJustReleased)
	- listen for key presses and record them into a string until Enter is pressed (take the typewriter example from ebiten demos as reference)
	- try to execute basic commands like "addplayer Alice", "removeplayer Bob", "listplayers", "startgame", "quit"
	- all commands and their outputs are added to a list of strings.
	- the most recent N strings are rendered to the screen as a simple console log
- config loading (using go-toml)
- game backend (game session, game state, input handling & abstraction, map system) and at the same time a simple renderer that can display the game state using the vector graphics lib from ebiten
	- this includes implementing basic snake movement, apple & item spawning and collection, snake growth/shrinking, and death/respawn logic, item entities (bullets, bots, etc), collision detection
	- an input system that can read keyboard input and map it to player actions
	- implement map loading from JSON files
	- all visible elements define a draw function that renders them to the screen. In this step this can be very basic shapes (rectangles, circles, lines)
- resource manager and asset loading
- sound effects and music playback
- statistics tracking
- GUI overlays (main menu, pause menu, HUD, post-match stats)
