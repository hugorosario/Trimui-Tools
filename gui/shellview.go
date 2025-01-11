package gui

import (
	"strings"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type ShellView struct {
	renderer     *sdl.Renderer
	lineHeight   int
	maxLineWidth int
	charWidth    int
	font         *ttf.Font
	color        *sdl.Color
	bgColor      *sdl.Color
	lines        []string
	position     sdl.Rect
	yOffset      int
	padding      int32
}

func NewShellView(renderer *sdl.Renderer, position sdl.Rect, font *ttf.Font, color sdl.Color, bgColor sdl.Color, padding int32) *ShellView {
	return &ShellView{
		renderer:     renderer,
		position:     position,
		font:         font,
		color:        &color,
		bgColor:      &bgColor,
		lineHeight:   TextHeight(font, "A"),
		charWidth:    TextWidth(font, "A"),
		maxLineWidth: int(position.W),
		padding:      padding,
	}
}

func (t *ShellView) SetContent(text string) {
	t.Clear()
	t.AddText(text)
}

func (t *ShellView) parseLines(lines []string) []string {
	var parsedLines []string
	for _, line := range lines {
		if len(line) > 0 {
			parsedLines = append(parsedLines, WrapText(line, t.charWidth, t.maxLineWidth)...)
		}
	}
	return parsedLines
}

func (t *ShellView) AddText(text string) {
	atBottom := t.AtBottom()
	t.lines = append(t.lines, t.parseLines(strings.Split(text, "\n"))...)
	if atBottom {
		t.GoToBottom()
	}
}

func (t ShellView) maxYOffset() int {
	return max(0, len(t.lines)-int(t.position.H)/t.lineHeight)
}

func (t *ShellView) SetYOffset(n int) {
	t.yOffset = clamp(n, 0, t.maxYOffset())
}

func (t ShellView) AtTop() bool {
	return t.yOffset <= 0
}

func (t ShellView) AtBottom() bool {
	return (t.yOffset >= t.maxYOffset())
}

func (t *ShellView) ScrollDown(n int) {
	t.SetYOffset(t.yOffset + n)
}

func (t *ShellView) ScrollUp(n int) {
	t.SetYOffset(t.yOffset - n)
}

func (t *ShellView) GoToBottom() {
	t.SetYOffset(t.maxYOffset())
}

func (t ShellView) visibleLines() (lines []string) {
	if len(t.lines) > 0 {
		top := max(0, t.yOffset)
		bottom := min(len(t.lines), t.yOffset+int(t.position.H)/t.lineHeight)
		lines = t.lines[top:bottom]
	}
	return lines
}

func (t *ShellView) Draw() {
	//render view background
	r, g, b, a, _ := t.renderer.GetDrawColor()
	_ = t.renderer.SetDrawColor(t.bgColor.R, t.bgColor.G, t.bgColor.B, t.bgColor.A)
	_ = t.renderer.FillRect(&sdl.Rect{X: t.position.X, Y: t.position.Y, W: t.position.W, H: t.position.H})
	_ = t.renderer.SetDrawColor(r, g, b, a)

	for index, item := range t.visibleLines() {
		if len(item) == 0 {
			continue
		}
		DrawTextCenter(t.renderer, item, sdl.Rect{X: t.position.X + t.padding, Y: t.position.Y + t.padding + int32(t.lineHeight)*int32(index), W: t.position.W - t.padding, H: int32(t.lineHeight) - t.padding}, *t.color, t.font, false, false)
	}
}

func (t *ShellView) GetScrollOffset() int {
	return t.yOffset
}

func (t *ShellView) GetText() []string {
	return t.lines
}

func (t *ShellView) SetText(text string) {
	t.Clear()
	lines := strings.Split(text, "\n")
	t.lines = t.parseLines(lines)
	t.GoToBottom()
}

func (t *ShellView) Clear() {
	t.lines = []string{}
	t.SetYOffset(0)
}
