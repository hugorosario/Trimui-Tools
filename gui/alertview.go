package gui

import (
	"strings"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type AlertView struct {
	renderer     *sdl.Renderer
	lineHeight   int
	maxLineWidth int
	charWidth    int
	font         *ttf.Font
	color        *sdl.Color
	bgColor      *sdl.Color
	lines        []string
	position     sdl.Rect
	parsedText   string
}

func NewAlertView(renderer *sdl.Renderer, position sdl.Rect, font *ttf.Font, color sdl.Color, bgColor sdl.Color) *AlertView {
	return &AlertView{
		renderer:     renderer,
		position:     position,
		font:         font,
		color:        &color,
		bgColor:      &bgColor,
		lineHeight:   TextHeight(font, "A"),
		charWidth:    TextWidth(font, "A"),
		maxLineWidth: int(position.W),
	}
}

func (t *AlertView) GetText() string {
	return t.parsedText
}

func (t *AlertView) SetText(text string) {
	lines := strings.Split(text, "\n")
	t.lines = t.parseLines(lines)
	t.parsedText = strings.Join(t.lines, "\n")
}

func (t *AlertView) parseLines(lines []string) []string {
	var parsedLines []string
	for _, line := range lines {
		if len(line) > 0 {
			parsedLines = append(parsedLines, WrapText(line, t.charWidth, t.maxLineWidth)...)
		}
	}
	return parsedLines
}

func (t *AlertView) Draw() {
	//render view background
	r, g, b, a, _ := t.renderer.GetDrawColor()
	_ = t.renderer.SetDrawColor(t.bgColor.R, t.bgColor.G, t.bgColor.B, t.bgColor.A)
	_ = t.renderer.FillRect(&sdl.Rect{X: t.position.X, Y: t.position.Y, W: t.position.W, H: t.position.H})
	_ = t.renderer.SetDrawColor(r, g, b, a)
	yPos := t.position.Y + (t.position.H-int32(len(t.lines)*t.lineHeight))/2
	for _, line := range t.lines {
		DrawTextCenter(t.renderer, line, sdl.Rect{X: t.position.X, Y: yPos, W: t.position.W, H: int32(t.lineHeight)}, *t.color, t.font, true, true)
		yPos += int32(t.lineHeight)
	}
}
