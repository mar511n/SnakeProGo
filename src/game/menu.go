package game

import (
	"fmt"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type MainMenu struct {
	CurrentContext Context
	Toaster        *Toaster
	mainContext    *MainMenuContext
	replayContext  *ReplayContext
	history        *HistoryData
	replay         *ReplaySession
	resources      *ResourceManager
	renderer       Renderer
	OnStartGame    func()
}

func NewMainMenu(startGameCallback func(), resources *ResourceManager) *MainMenu {
	mm := &MainMenu{
		OnStartGame:   startGameCallback,
		mainContext:   &MainMenuContext{},
		replayContext: &ReplayContext{},
		resources:     resources,
	}
	mm.Toaster = &Toaster{
		Toasts:          []*Toast{},
		Pos:             Vec2f{X: 0.02 * GConfig.AspectRatio(), Y: 0.98},
		ToastHeight:     0.05,
		ToastSpacing:    0.008,
		textOffset:      0.009,
		TextColor:       color.White,
		BoxColor:        color.RGBA{0, 0, 0, 150},
		BoxBorderColor:  color.White,
		BoxBorderWidth:  3,
		HorizontalAlign: ToastAlignRight,
		VerticalAlign:   ToastAlignUp,
	}
	mm.Toaster.Initialize(resources)
	mm.mainContext.Initialize(mm, resources)
	mm.replayContext.Initialize(mm)
	mm.CurrentContext = mm.mainContext

	mm.renderer = NewDefaultRenderer(mm.resources)
	return mm
}

func (m *MainMenu) Update() error {
	m.Toaster.Update()
	if m.CurrentContext != nil {
		return m.CurrentContext.Update()
	}
	FatalError("MainMenu: No current context to update.")
	return nil
}

func (m *MainMenu) Draw(screen *ebiten.Image) {
	if m.CurrentContext != nil {
		m.CurrentContext.Draw(screen)
		m.Toaster.Draw(screen)
	} else {
		FatalError("MainMenu: No current context to draw.")
	}
}

func (m *MainMenu) OnGameSessionDone(winners []string, hist *HistoryData, resources *ResourceManager) {
	m.replay = NewReplaySession(hist, resources)
	m.history = hist
	m.renderer.InitRender(true, m.replay.State)
	SavePlayerConfigs()
}

func (m *MainMenu) PrintMessage(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	LogInfo(format, args...)
	m.Toaster.AddToast(msg, 3.0)
}

func (m *MainMenu) AddPlayer(name string) {
	_, alreadyAdded := PConfigs[name]

	if alreadyAdded {
		m.PrintMessage("Player %s is already in the game.", name)
	} else {
		_, loaded := GetPlayerConfig(name) // Ensure config is loaded/created
		m.mainContext.AddPlayerName(name, m)
		if loaded {
			m.PrintMessage("Config loaded and player %s added.", name)
		} else {
			m.PrintMessage("Config created and player %s added.", name)
		}
	}
}

func (m *MainMenu) RemovePlayer(name string) {
	_, found := PConfigs[name]
	if found {
		delete(PConfigs, name)
		m.mainContext.RemovePlayerName(name, m)
		m.PrintMessage("Removed player: %s", name)
	} else {
		m.PrintMessage("Player not found: %s", name)
	}
}

func (m *MainMenu) ShowControllers() {
	ids := ebiten.AppendGamepadIDs([]ebiten.GamepadID{})
	sdlids := make([]string, len(ids))
	for i, id := range ids {
		sdlids[i] = ebiten.GamepadSDLID(id)
	}
	if len(sdlids) == 0 {
		m.PrintMessage("No controllers detected.")
	} else {
		m.PrintMessage("Detected controllers: %v", sdlids)
	}
	for playerName, pconfig := range PConfigs {
		m.PrintMessage("Player: %s, Config: %+v", playerName, pconfig.KeyMap)
	}
}

func (m *MainMenu) Quit() {
	m.PrintMessage("Quitting...")
	os.Exit(0)
}

/*
	switch command {
	case "showconfig", "sc":
		if len(args) < 1 {
			m.AddHistory("Usage: showconfig [global/game/<username>]")
		} else {
			target := args[0]
			var cfg interface{}
			found := true

			switch target {
			case "global":
				cfg = GConfig
			case "game":
				cfg = GPConfig
			default:
				// Assume player name
				if _, ok := PConfigs[target]; ok {
					cfg = PConfigs[target]
				} else {
					found = false
					m.AddHistory("Config not found for: %s. (Is player added?)", target)
				}
			}

			if found {
				data, err := toml.Marshal(cfg)
				if err != nil {
					m.AddHistory("Error displaying config: %v", err)
				} else {
					// Split lines so they show up nicely in history
					lines := strings.Split(string(data), "\n")
					m.AddHistory("--- Config: %s ---", target)
					m.history = append(m.history, lines...)
					m.AddHistory("----------------")
				}
			}
		}
} */

type MainMenuContext struct {
	menu               *MainMenu
	resources          *ResourceManager
	guiMain            *GuiContext
	background         *ebiten.Image
	newplayertextedit  *GuiEditText
	playerlist         map[string]int
	playernamefontsize float64
	static_elements    []GuiElement
}

func (c *MainMenuContext) Initialize(menu *MainMenu, resources *ResourceManager) {
	c.menu = menu
	c.resources = resources
	c.background = resources.Images["UI"]["WinBack"]
	bw := 0.045
	xspace := 0.004
	x0 := 1.0 - (bw + xspace)
	y0 := 0.13
	c.playernamefontsize = GetFontSizeForHeight(c.resources.Fonts["default"], 0.037)
	elements := make([]GuiElement, 0)
	startButton := &GuiButton{
		Rect:  Rectf{Pos: Vec2f{X: x0, Y: y0}, Size: Vec2f{X: bw, Y: bw / GConfig.AspectRatio()}},
		Image: resources.Images["UI"]["play"],
		OnClick: func() {
			LogInfo("Starting game...")
			if c.menu.OnStartGame != nil {
				c.menu.OnStartGame()
			}
		},
		neighbors: make(map[GuiDirection]GuiElement),
	}
	replayBtn := &GuiButton{
		Rect:  Rectf{Pos: Vec2f{X: x0, Y: y0 + (bw+xspace)/GConfig.AspectRatio()}, Size: Vec2f{X: bw, Y: bw / GConfig.AspectRatio()}},
		Image: resources.Images["UI"]["replay"],
		OnClick: func() {
			if c.menu.replay != nil {
				c.menu.PrintMessage("Starting replay...")
				c.menu.mainContext.StartContextSwitch(c.menu.replayContext, func(toContext Context) { c.menu.CurrentContext = c.menu.replayContext })
			} else {
				c.menu.PrintMessage("No replay available.")
			}
		},
		neighbors: make(map[GuiDirection]GuiElement),
	}
	controllerBtn := &GuiButton{
		Rect:  Rectf{Pos: Vec2f{X: x0, Y: y0 + (bw+xspace)*2/GConfig.AspectRatio()}, Size: Vec2f{X: bw, Y: bw / GConfig.AspectRatio()}},
		Image: resources.Images["UI"]["controller"],
		OnClick: func() {
			c.menu.ShowControllers()
		},
		neighbors: make(map[GuiDirection]GuiElement),
	}
	quitButton := &GuiButton{
		Rect:  Rectf{Pos: Vec2f{X: x0, Y: y0 + (bw+xspace)*3/GConfig.AspectRatio()}, Size: Vec2f{X: bw, Y: bw / GConfig.AspectRatio()}},
		Image: resources.Images["UI"]["quit"],
		OnClick: func() {
			c.menu.Quit()
		},
		neighbors: make(map[GuiDirection]GuiElement),
	}
	c.newplayertextedit = &GuiEditText{
		Rect:             Rectf{Pos: Vec2f{X: 0.37, Y: 0.45}, Size: Vec2f{X: 0.25, Y: 0.04}},
		neighbors:        make(map[GuiDirection]GuiElement),
		resources:        c.resources,
		AlloweLineBreaks: false,
		DisallowedChars:  "",
		TextColor:        color.Black,
		BoxColor:         color.RGBA{0, 0, 0, 50},
		BoxFocusColor:    color.RGBA{32, 32, 32, 50},
		BoxBorderColor:   color.Black,
		BoxBorderWidth:   3,
		MaxCharacters:    20,
		Text:             "[username]",
		OnSelectFunc: func(t *GuiEditText) {
			if t.Text == "[username]" {
				t.Text = ""
			}
		},
		OnDeselectFunc: func(t *GuiEditText) {
			if t.Text == "" {
				t.Text = "[username]"
			}
		},
	}
	c.newplayertextedit.FontSize = c.playernamefontsize
	addPlayerBtn := &GuiButton{
		Rect:  Rectf{Pos: Vec2f{X: c.newplayertextedit.Rect.Pos.X + c.newplayertextedit.Rect.Size.X + 0.015, Y: 0.45}, Size: Vec2f{X: 0.04 * GConfig.AspectRatio(), Y: 0.04}},
		Image: resources.Images["UI"]["add"],
		OnClick: func() {
			c.menu.AddPlayer(c.newplayertextedit.Text)
		},
		neighbors: make(map[GuiDirection]GuiElement),
	}
	subPlayerBtn := &GuiButton{
		Rect:  Rectf{Pos: Vec2f{X: addPlayerBtn.Rect.Pos.X + 0.04*GConfig.AspectRatio() + 0.005, Y: 0.45}, Size: Vec2f{X: 0.04 * GConfig.AspectRatio(), Y: 0.04}},
		Image: resources.Images["UI"]["sub"],
		OnClick: func() {
			c.menu.RemovePlayer(c.newplayertextedit.Text)
		},
		neighbors: make(map[GuiDirection]GuiElement),
	}

	elements = append(elements, c.newplayertextedit, addPlayerBtn, subPlayerBtn, startButton, replayBtn, controllerBtn, quitButton)
	for i, el := range elements {
		el.SetNeighboringElement(GuiDirDown, elements[(i+1)%len(elements)])
		el.SetNeighboringElement(GuiDirUp, elements[(i-1+len(elements))%len(elements)])
	}
	c.static_elements = elements

	c.guiMain = NewGuiContext(elements)
	c.guiMain.drawDebugRects = true
	for _, btn := range elements {
		if b, ok := btn.(*GuiButton); ok {
			b.deselect = func() bool {
				c.guiMain.selectedElement = -1
				return true
			}
		}
	}
	c.playerlist = make(map[string]int)
	for name, _ := range PConfigs {
		c.AddPlayerName(name, menu)
	}
}

func (c *MainMenuContext) AddPlayerName(name string, m *MainMenu) bool {
	if _, ok := c.playerlist[name]; ok {
		return false
	}
	label := &GuiText{
		focused:        false,
		selected:       false,
		resources:      m.resources,
		neighbors:      make(map[GuiDirection]GuiElement),
		Rect:           Rectf{Pos: Vec2f{X: 0.37, Y: 0.51 + 0.045*float64(len(c.playerlist))}, Size: Vec2f{X: 0.25, Y: 0.04}},
		Text:           name,
		FontSize:       m.mainContext.playernamefontsize,
		TextColor:      color.Black,
		BoxColor:       color.RGBA{0, 0, 0, 0},
		BoxBorderColor: color.Transparent,
		BoxBorderWidth: 3,
		OnSelectFunc: func(t *GuiText) {
			m.mainContext.newplayertextedit.Text = t.Text
			t.OnDeselect()
			m.mainContext.guiMain.selectedElement = -1
		},
	}
	m.mainContext.guiMain.Elements = append(m.mainContext.guiMain.Elements, label)
	label.SetContextID(len(m.mainContext.guiMain.Elements) - 1)
	c.playerlist[name] = label.GetContextID()
	return true
}

func (c *MainMenuContext) RemovePlayerName(name string, m *MainMenu) {
	if _, ok := c.playerlist[name]; ok {
		m.mainContext.guiMain.Elements = c.static_elements
		plnames := make([]string, 0, len(c.playerlist)-1)
		for plname, _ := range c.playerlist {
			if plname != name {
				plnames = append(plnames, plname)
			}
		}
		c.playerlist = make(map[string]int)
		for _, plname := range plnames {
			c.AddPlayerName(plname, m)
		}
	}
}

func (c *MainMenuContext) Update() error {
	return c.guiMain.Update()
}
func (c *MainMenuContext) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(screen.Bounds().Dx())/float64(c.background.Bounds().Dx()), float64(screen.Bounds().Dy())/float64(c.background.Bounds().Dy()))
	screen.DrawImage(c.background, op)
	/*
		op2 := &text.DrawOptions{}
		op2.ColorScale.ScaleWithColor(color.Black)
		op2.LayoutOptions.LineSpacing = 0.04 * float64(GConfig.ScreenHeight)
		var plnames []string
		for name, _ := range PConfigs {
			plnames = append(plnames, name)
		}
		sort.Strings(plnames)
		DrawTextRelative(screen, strings.Join(plnames, "\n"), &text.GoTextFace{Source: c.resources.Fonts["default"], Size: c.playernamefontsize}, Vec2f{X: 0.37, Y: 0.51}, op2)
	*/
	c.guiMain.Draw(screen)
}
func (c *MainMenuContext) StartContextSwitch(newContext Context, switchContext func(toContext Context)) error {
	switchContext(newContext)
	return nil
}

type ReplayContext struct {
	menu *MainMenu
}

func (r *ReplayContext) StartContextSwitch(newContext Context, switchContext func(toContext Context)) error {
	switchContext(newContext)
	return nil
}

func (r *ReplayContext) Initialize(menu *MainMenu) {
	r.menu = menu
}

func (r *ReplayContext) Update() error {
	if r.menu.replay != nil {
		r.menu.replay.Update()
		if r.menu.replay.IsFinished() {
			r.menu.PrintMessage("Replay finished.")
			r.menu.replay = NewReplaySession(r.menu.history, r.menu.resources)
			r.StartContextSwitch(r.menu.mainContext, func(toContext Context) { r.menu.CurrentContext = r.menu.mainContext })
		}
	}
	return nil
}

func (r *ReplayContext) Draw(screen *ebiten.Image) {
	if r.menu.replay != nil && r.menu.replay.State != nil && r.menu.renderer != nil {
		r.menu.renderer.Render(r.menu.replay.State, screen)
	}
}
