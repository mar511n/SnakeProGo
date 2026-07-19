package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type ToastAlignmentVertical int

const (
	ToastAlignUp ToastAlignmentVertical = iota
	ToastAlignDown
)

type ToastAlignmentHorizontal int

const (
	ToastAlignLeft ToastAlignmentHorizontal = iota
	ToastAlignCenter
	ToastAlignRight
)

type Toast struct {
	Message string
	Ticks   int
}

func (t *Toast) Draw(screen *ebiten.Image, tcol, bcol, bbcol color.Color, bbwidth float64, font *text.GoTextFace, Rect Rectf, textoff Vec2f) {
	DrawRectRelative(screen, Rect.Pos, Rect.Size, bcol, bbcol, bbwidth)
	op := &text.DrawOptions{}
	op.ColorScale.ScaleWithColor(tcol)
	DrawTextRelative(
		screen,
		t.Message,
		font,
		Rect.Pos.Add(textoff),
		op,
	)
}

type Toaster struct {
	Toasts          []*Toast
	Pos             Vec2f
	ToastHeight     float64
	ToastSpacing    float64
	textOffset      float64
	fontSize        float64
	TextColor       color.Color
	BoxColor        color.Color
	BoxBorderColor  color.Color
	BoxBorderWidth  float64
	HorizontalAlign ToastAlignmentHorizontal // can be left/center/right
	VerticalAlign   ToastAlignmentVertical   // can be up/down
	resources       *ResourceManager
}

func (t *Toaster) Initialize(resources *ResourceManager) {
	t.resources = resources
	t.fontSize = GetFontSizeForHeight(t.resources.Fonts["default"], t.ToastHeight-t.textOffset*2)
}

func (t *Toaster) AddToast(message string, duration float64) {
	durationTicks := int(duration * float64(GConfig.TPS))
	t.Toasts = append(t.Toasts, &Toast{Message: message, Ticks: durationTicks})
}

func (t *Toaster) Update() error {
	for i := len(t.Toasts) - 1; i >= 0; i-- {
		toast := t.Toasts[i]
		toast.Ticks--
		if toast.Ticks <= 0 {
			t.Toasts = append(t.Toasts[:i], t.Toasts[i+1:]...)
		}
	}
	return nil
}

func (t *Toaster) Draw(screen *ebiten.Image) {
	font := &text.GoTextFace{
		Source: t.resources.Fonts["default"],
		Size:   t.fontSize,
	}
	for i, toast := range t.Toasts {
		tmw, tmh := text.Measure(toast.Message, font, 0.0)
		tmw /= float64(GConfig.ScreenWidth)
		tmh /= float64(GConfig.ScreenHeight)
		tmh += t.textOffset * 2
		anchorPos := Vec2f{X: t.Pos.X, Y: t.Pos.Y}
		if t.VerticalAlign == ToastAlignUp {
			anchorPos.Y -= float64(i)*(tmh+t.ToastSpacing) + tmh
		} else {
			anchorPos.Y += float64(i) * (tmh + t.ToastSpacing)
		}
		if t.HorizontalAlign == ToastAlignCenter {
			anchorPos.X -= tmw/2 + t.textOffset*GConfig.AspectRatio()
		} else if t.HorizontalAlign == ToastAlignLeft {
			anchorPos.X -= tmw + t.textOffset*2*GConfig.AspectRatio()
		}
		toastRect := Rectf{Pos: anchorPos, Size: Vec2f{X: tmw + t.textOffset*2*GConfig.AspectRatio(), Y: tmh}}
		toast.Draw(screen, t.TextColor, t.BoxColor, t.BoxBorderColor, t.BoxBorderWidth, font, toastRect, Vec2f{X: t.textOffset * GConfig.AspectRatio(), Y: t.textOffset})
	}
}
