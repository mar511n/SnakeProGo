package game

import (
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type GuiAction int

const (
	GuiActionNone GuiAction = iota
	GuiActionUp
	GuiActionDown
	GuiActionLeft
	GuiActionRight
	GuiActionSelect
	GuiActionDeselect
)

type GuiDirection int

const (
	GuiDirUp GuiDirection = iota
	GuiDirDown
	GuiDirLeft
	GuiDirRight
)

// Renders and handles input to a list of GUI elements
type GuiContext struct {
	Elements        []GuiElement
	focusedElement  int
	selectedElement int
	inputframe      *InputFrame
	inputconfigs    map[int]*PlayerConfig
	lastMousePos    Vec2i
	initialized     bool
	drawDebugRects  bool
}

func NewGuiContext(elements []GuiElement) *GuiContext {
	gc := &GuiContext{
		Elements:        elements,
		focusedElement:  0,
		selectedElement: -1,
		inputframe:      &InputFrame{},
		inputconfigs:    make(map[int]*PlayerConfig),
		lastMousePos:    Vec2i{0, 0},
		drawDebugRects:  false,
	}
	for i, elem := range gc.Elements {
		elem.SetContextID(i)
	}
	pli := 0
	for _, plcfg := range PConfigs {
		gc.inputconfigs[pli] = plcfg
		pli++
	}
	gc.inputconfigs[pli] = &PlayerConfig{
		Name: "DefaultWASD",
		KeyMap: map[string]string{
			"up":         ebiten.KeyW.String(),
			"down":       ebiten.KeyS.String(),
			"left":       ebiten.KeyA.String(),
			"right":      ebiten.KeyD.String(),
			"turn_left":  "",
			"turn_right": "",
			"use_item":   ebiten.KeySpace.String(),
		},
		ControllerMap:   map[string]string{},
		ControllerSDLID: "",
		Stats:           make(map[string]int),
	}
	pli++
	gc.inputconfigs[pli] = &PlayerConfig{
		Name: "DefaultArrows",
		KeyMap: map[string]string{
			"up":         ebiten.KeyUp.String(),
			"down":       ebiten.KeyDown.String(),
			"left":       ebiten.KeyLeft.String(),
			"right":      ebiten.KeyRight.String(),
			"turn_left":  "",
			"turn_right": "",
			"use_item":   ebiten.KeyEnter.String(),
		},
		ControllerMap:   map[string]string{},
		ControllerSDLID: "",
		Stats:           make(map[string]int),
	}
	return gc
}

func (c *GuiContext) GetMergedInputframe() (dirAction, selectAction GuiAction) {
	dirAction = GuiActionNone
	selectAction = GuiActionNone
	c.inputframe.Process(c.inputconfigs)
	for id, action := range c.inputframe.Directions {
		if dirAction == GuiActionNone || c.inputconfigs[id].Name == "DefaultWASD" || c.inputconfigs[id].Name == "DefaultArrows" {
			switch action {
			case ActionUp:
				dirAction = GuiActionUp
			case ActionDown:
				dirAction = GuiActionDown
			case ActionLeft:
				dirAction = GuiActionLeft
			case ActionRight:
				dirAction = GuiActionRight
			}
		}
	}
	for _, used := range c.inputframe.ItemsUsed {
		if used {
			if c.selectedElement == -1 {
				selectAction = GuiActionSelect
			} else {
				selectAction = GuiActionDeselect
			}
			break
		}
	}
	return
}

func (c *GuiContext) MoveFocus(newElement int) {
	if newElement >= 0 && newElement < len(c.Elements) && newElement != c.focusedElement {
		c.Elements[c.focusedElement].OnFocusLost()
		c.focusedElement = newElement
		c.Elements[c.focusedElement].OnFocusGained()
	}
}

func (c *GuiContext) SelectElement(elementID int) {
	if elementID >= 0 && elementID < len(c.Elements) && elementID != c.selectedElement {
		c.selectedElement = elementID
		c.Elements[c.selectedElement].OnSelect()
	}
}

func (c *GuiContext) TryToFindNeighboringElement(elementID int, dir GuiDirection) (GuiElement, bool) {
	if elementID >= 0 && elementID < len(c.Elements) {
		el, found := c.Elements[elementID].GetNeighboringElement(dir)
		if !found {
			el, found = c.Elements[elementID].GetNeighboringElement((dir + 2) % 4)
		}
		return el, found
	}
	return nil, false
}

func (c *GuiContext) Update() error {
	if !c.initialized {
		c.Elements[c.focusedElement].OnFocusGained()
		c.initialized = true
	}
	dirAction, selectAction := c.GetMergedInputframe()

	leftClick := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	nX, nY := ebiten.CursorPosition()
	mouseMove := c.lastMousePos.X != int16(nX) || c.lastMousePos.Y != int16(nY)
	mouseElement := -1
	for _, elem := range c.Elements {
		if elem.Bounds().Contains(Vec2f{float64(nX) / float64(GConfig.ScreenWidth), float64(nY) / float64(GConfig.ScreenHeight)}) {
			mouseElement = elem.GetContextID()
			break
		}
	}

	if c.selectedElement == -1 {
		if selectAction == GuiActionSelect {
			c.SelectElement(c.focusedElement)
		} else if mouseElement != -1 && leftClick {
			c.SelectElement(mouseElement)
		}
		if dirAction != GuiActionNone {
			neighbor, found := c.TryToFindNeighboringElement(c.focusedElement, GuiDirection(dirAction-1))
			if found {
				c.MoveFocus(neighbor.GetContextID())
			}
		} else if mouseElement != -1 && mouseMove {
			c.MoveFocus(mouseElement)
		}
	} else {
		if selectAction == GuiActionDeselect || (leftClick && mouseElement != c.selectedElement) {
			c.Elements[c.selectedElement].OnDeselect()
			c.selectedElement = -1
		} else {
			c.Elements[c.selectedElement].OnInput(dirAction, selectAction)
		}
	}
	for _, elem := range c.Elements {
		elem.Update()
	}
	c.lastMousePos = Vec2i{int16(nX), int16(nY)}
	return nil
}

func (c *GuiContext) Draw(screen *ebiten.Image) {
	for _, elem := range c.Elements {
		if c.drawDebugRects {
			DrawRectRelative(screen, elem.Bounds().Pos, elem.Bounds().Size, color.Transparent, color.RGBA{255, 0, 255, 128}, 3)
		}
		elem.Draw(screen)
	}
}

func DrawImageRelative(screen *ebiten.Image, img *ebiten.Image, pos, size Vec2f, op *ebiten.DrawImageOptions) {
	sw := float64(screen.Bounds().Dx())
	sh := float64(screen.Bounds().Dy())
	x0 := sw * pos.X
	y0 := sh * pos.Y
	w := sw * size.X
	h := sh * size.Y
	op.GeoM.Scale(w/float64(img.Bounds().Dx()), h/float64(img.Bounds().Dy()))
	op.GeoM.Translate(x0, y0)
	screen.DrawImage(img, op)
}

func DrawTextRelative(screen *ebiten.Image, txt string, face text.Face, pos Vec2f, op *text.DrawOptions) {
	x0 := float64(screen.Bounds().Dx()) * pos.X
	y0 := float64(screen.Bounds().Dy()) * pos.Y
	op.GeoM.Translate(x0, y0)
	text.Draw(screen, txt, face, op)
}

func DrawRectRelative(screen *ebiten.Image, pos, size Vec2f, fillColor, strokeColor color.Color, strokeWidth float64) {
	x0 := float32(float64(screen.Bounds().Dx()) * pos.X)
	y0 := float32(float64(screen.Bounds().Dy()) * pos.Y)
	w := float32(float64(screen.Bounds().Dx()) * size.X)
	h := float32(float64(screen.Bounds().Dy()) * size.Y)
	vector.FillRect(screen, x0, y0, w, h, fillColor, true)
	vector.StrokeRect(screen, x0, y0, w, h, float32(strokeWidth), strokeColor, true)
}

func GetFontSizeForHeight(source *text.GoTextFaceSource, targetHeight float64) float64 {
	testSize := targetHeight * float64(GConfig.ScreenWidth)
	for i := 0; i < 50; i++ {
		testFace := &text.GoTextFace{Source: source, Size: testSize}
		height := testFace.Metrics().CapHeight * float64(GConfig.ScreenWidth)
		if math.Abs(height-targetHeight) < 1 {
			return testSize
		}
		if height < targetHeight {
			testSize += 0.5
		} else {
			testSize -= 0.5
		}
	}
	return testSize
}

type GuiButton struct {
	Rect      Rectf
	Image     *ebiten.Image
	OnClick   func()
	focused   bool
	selected  bool
	contextID int
	deselect  func() bool
	neighbors map[GuiDirection]GuiElement
}

func (b *GuiButton) Bounds() *Rectf      { return &b.Rect }
func (b *GuiButton) GetContextID() int   { return b.contextID }
func (b *GuiButton) SetContextID(id int) { b.contextID = id }
func (b *GuiButton) OnFocusGained()      { b.focused = true }
func (b *GuiButton) OnFocusLost()        { b.focused = false }
func (b *GuiButton) OnDeselect()         { b.selected = false }
func (b *GuiButton) Update() error       { return nil }
func (b *GuiButton) SetNeighboringElement(dir GuiDirection, element GuiElement) {
	b.neighbors[dir] = element
}
func (b *GuiButton) GetNeighboringElement(dir GuiDirection) (GuiElement, bool) {
	neighbor, found := b.neighbors[dir]
	return neighbor, found
}

func (b *GuiButton) OnSelect() {
	b.selected = true
	if b.OnClick != nil {
		b.OnClick()
		if b.deselect() {
			b.OnDeselect()
		}
	}
}

func (b *GuiButton) OnInput(dirAction GuiAction, selectAction GuiAction) {
	//if selectAction == GuiActionSelect || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
	//	b.OnClick()
	//}
}

func (b *GuiButton) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	if b.focused {
		op.ColorScale.Scale(1.2, 1.2, 1.2, 1.0)
	}
	DrawImageRelative(screen, b.Image, b.Rect.Pos, b.Rect.Size, op)
}

type GuiEditText struct {
	Rect             Rectf
	focused          bool
	selected         bool
	contextID        int
	blinkCounter     int
	runes            []rune
	Text             string
	neighbors        map[GuiDirection]GuiElement
	resources        *ResourceManager
	FontSize         float64
	TextColor        color.Color
	BoxColor         color.Color
	BoxFocusColor    color.Color
	BoxBorderColor   color.Color
	BoxBorderWidth   float64
	DisallowedChars  string
	AlloweLineBreaks bool
	MaxCharacters    int
	OnSelectFunc     func(t *GuiEditText)
	OnDeselectFunc   func(t *GuiEditText)
}

func (t *GuiEditText) Bounds() *Rectf      { return &t.Rect }
func (t *GuiEditText) GetContextID() int   { return t.contextID }
func (t *GuiEditText) SetContextID(id int) { t.contextID = id }
func (t *GuiEditText) OnFocusGained()      { t.focused = true }
func (t *GuiEditText) OnFocusLost()        { t.focused = false }
func (t *GuiEditText) OnSelect() {
	t.selected = true
	if t.OnSelectFunc != nil {
		t.OnSelectFunc(t)
	}
}
func (t *GuiEditText) OnDeselect() {
	t.selected = false
	t.blinkCounter = 0
	if t.OnDeselectFunc != nil {
		t.OnDeselectFunc(t)
	}
}
func (t *GuiEditText) Update() error { return nil }
func (t *GuiEditText) SetNeighboringElement(dir GuiDirection, element GuiElement) {
	t.neighbors[dir] = element
}
func (t *GuiEditText) GetNeighboringElement(dir GuiDirection) (GuiElement, bool) {
	neighbor, found := t.neighbors[dir]
	return neighbor, found
}

// repeatingKeyPressed return true when key is pressed considering the repeat state.
func repeatingKeyPressed(key ebiten.Key) bool {
	const (
		delay    = 30
		interval = 3
	)
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true
	}
	if d >= delay && (d-delay)%interval == 0 {
		return true
	}
	return false
}

func (t *GuiEditText) OnInput(dirAction GuiAction, selectAction GuiAction) {
	t.runes = ebiten.AppendInputChars(t.runes[:0])
	for _, r := range t.runes {
		if !strings.ContainsRune(t.DisallowedChars, r) {
			t.Text += string(r)
		}
	}

	// If the enter key is pressed, add a line break.
	if t.AlloweLineBreaks && (repeatingKeyPressed(ebiten.KeyEnter) || repeatingKeyPressed(ebiten.KeyNumpadEnter)) {
		t.Text += "\n"
	}

	if t.MaxCharacters > 0 && len(t.Text) > t.MaxCharacters {
		t.Text = t.Text[:t.MaxCharacters]
	}

	// If the backspace key is pressed, remove one character.
	if repeatingKeyPressed(ebiten.KeyBackspace) {
		if len(t.Text) >= 1 {
			t.Text = t.Text[:len(t.Text)-1]
		}
	}

	t.blinkCounter++
}

func (t *GuiEditText) Draw(screen *ebiten.Image) {
	et := t.Text
	if t.blinkCounter%60 > 30 {
		et += "_"
	}
	fillcol := t.BoxColor
	if t.focused {
		fillcol = t.BoxFocusColor
	}
	DrawRectRelative(screen, t.Rect.Pos.Sub(Vec2f{0.01 * GConfig.AspectRatio(), 0.01}), t.Rect.Size.Add(Vec2f{0.02 * GConfig.AspectRatio(), 0.02}), fillcol, t.BoxBorderColor, t.BoxBorderWidth)
	op := &text.DrawOptions{}
	op.ColorScale.ScaleWithColor(t.TextColor)
	DrawTextRelative(
		screen,
		et,
		&text.GoTextFace{
			Source: t.resources.Fonts["comic"],
			Size:   t.FontSize,
		},
		t.Rect.Pos,
		op,
	)
}
