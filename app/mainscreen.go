package app

import (
	"context"
	"fmt"
	"os"

	"github.com/hugorosario/trimuitools/shell"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	itemHeight   = 120
	itemSpacing  = 8
	visibleItems = 5
)

type OptionsList struct {
	Title string
	List  *List[*menuItem]
}

type MainScreen struct {
	renderer       *sdl.Renderer
	shellView      *ShellView
	errorTextView  *AlertView
	lineView       *AlertView
	dialogView     *AlertView
	lists          []OptionsList
	initialized    bool
	title          string
	showShellView  bool
	shellRunnning  bool
	showLineView   bool
	showDialogView bool
	shellInput     chan string
	shellCancel    context.CancelFunc
}

func NewMainScreen(renderer *sdl.Renderer) (*MainScreen, error) {
	title, err := readJsonFileProperty("config.json", "label")
	if err != nil {
		title = "System Tools"
	}
	screen := &MainScreen{
		title:    title,
		renderer: renderer,
		errorTextView: NewAlertView(
			renderer,
			sdl.Rect{X: 0, Y: 75, W: DisplayWidth, H: DisplayHeight - 125},
			ContentFont1,
			ContentColor1,
			HexToColor("#9AE4080A"),
		),
		lineView: NewAlertView(
			renderer,
			sdl.Rect{X: 0, Y: 70, W: DisplayWidth, H: DisplayHeight - 120},
			ContentFont1,
			ContentColor1,
			HexToColor("#9A000000"),
		),
		dialogView: NewAlertView(
			renderer,
			sdl.Rect{X: 0, Y: 70, W: DisplayWidth, H: DisplayHeight - 120},
			ContentFont1,
			ContentColor1,
			HexToColor("#00FFFFFF"),
		),
		shellView: NewShellView(
			renderer,
			sdl.Rect{X: 0, Y: 70, W: DisplayWidth, H: DisplayHeight - 120},
		),
	}
	listPosition := sdl.Point{X: 20, Y: 78}
	mainlist := NewList(
		renderer,
		int(visibleItems),
		listPosition,
		func(index int, item Item[*menuItem], selected bool) {
			itemRect := sdl.Rect{X: listPosition.X, Y: listPosition.Y + (itemHeight+itemSpacing)*int32(index), W: 984, H: itemHeight}
			screen.RenderListItem(renderer, item.Value, itemRect, selected)
		},
	)
	screen.lists = make([]OptionsList, 0)
	screen.lists = append(screen.lists, OptionsList{
		Title: screen.title,
		List:  mainlist,
	})

	return screen, nil
}

func (h *MainScreen) InitMenu() {
	if h.initialized {
		return
	}
	_ = h.renderer.SetDrawColor(0, 0, 0, 255)
	DrawTextCenter(h.renderer, "Loading...", sdl.Rect{X: 0, Y: 0, W: DisplayWidth, H: DisplayHeight}, ContentColor1, ContentFont1, true, true)
	h.renderer.Present()

	//execute the ./init.sh script if it exists
	_, err := os.Stat("./init.sh")
	if err == nil {
		_, _ = shell.RunCommandSync("./init.sh")
	}

	items, err := h.loadItems()
	if err != nil {
		h.SetError(fmt.Sprintf("Error loading menu items.\n%s", err))
	} else {
		h.title = h.lists[len(h.lists)-1].Title
		h.GetList().SetItems(items)
	}
	h.initialized = true
}

func (h *MainScreen) HandleInput(event InputEvent) {
	if !h.initialized {
		return
	}
	switch event.KeyCode {
	case "DOWN":
		if h.InError() {
			return
		}
		if h.InShell() {
			if h.showShellView {
				h.shellView.ScrollDown(1)
			}
			return
		} else {
			h.GetList().ScrollDown()
		}
	case "UP":
		if h.InError() {
			return
		}
		if h.InShell() {
			if h.showShellView {
				h.shellView.ScrollUp(1)
			}
			return
		} else {
			h.GetList().ScrollUp()
		}
	case "B":
		if h.InShell() {
			if h.IsExecuting() {
				if h.InDialog() {
					h.showDialogView = false
					if h.shellInput != nil {
						h.shellInput <- event.KeyCode + "\n"
					}
				} else if h.shellCancel != nil {
					h.shellCancel()
				}
			} else {
				h.title = h.lists[len(h.lists)-1].Title
				h.showShellView = false
				h.showLineView = false
			}
			return
		}

		if h.InError() {
			Close()
			return
		}

		if len(h.lists) > 1 {
			h.lists = h.lists[:len(h.lists)-1]
			h.title = h.lists[len(h.lists)-1].Title
		} else {
			Close()
		}
	case "LEFT":
		if h.InShell() || h.InError() {
			return
		}
		h.ChangeSelection(h.GetList().SelectedItem(), "LEFT")
	case "RIGHT":
		if h.InShell() || h.InError() {
			return
		}
		h.ChangeSelection(h.GetList().SelectedItem(), "RIGHT")
	case "A":
		if h.InError() {
			return
		}

		if h.InShell() {
			if h.InDialog() && h.IsExecuting() {
				h.showDialogView = false
				if h.shellInput != nil {
					h.shellInput <- event.KeyCode + "\n"
				}
			}
			return
		}

		h.DoItemAction(h.GetList().SelectedItem())
	case "SELECT":
		if h.InError() {
			return
		}
		if len(h.shellView.GetText()) > 0 {
			h.showLineView = false
			h.showShellView = true
		}
	}
}

func (h *MainScreen) GetList() *List[*menuItem] {
	return h.lists[len(h.lists)-1].List
}

func (h *MainScreen) GetTitle() string {
	return h.title
}

func (h *MainScreen) InError() bool {
	return len(h.errorTextView.GetText()) > 0
}

func (h *MainScreen) SetError(errorMsg string) {
	h.errorTextView.SetText(errorMsg)
}

func (h *MainScreen) InShell() bool {
	return h.showShellView || h.showLineView || h.showDialogView
}

func (h *MainScreen) IsExecuting() bool {
	return h.shellRunnning
}

func (h *MainScreen) InDialog() bool {
	return h.showDialogView
}

func (h *MainScreen) ChangeSelection(item *menuItem, direction string) {
	if item.Disabled {
		return
	}
	switch item.Type {
	case "select":
		if direction == "RIGHT" {
			if item.Selected < len(item.SelectItems)-1 {
				item.Selected++
			}
		}

		if direction == "LEFT" {
			if item.Selected > 0 {
				item.Selected--
			}
		}
	}
}

func (h *MainScreen) DrawThemeTextures() {
	DrawTexture(h.renderer, BackgroundTexture, sdl.Rect{X: 0, Y: 0, W: DisplayWidth, H: DisplayHeight})
	DrawTexture(h.renderer, TitleBarTexture, sdl.Rect{X: 0, Y: 0, W: DisplayWidth, H: 70})
	DrawTexture(h.renderer, TipsBarTexture, sdl.Rect{X: 0, Y: DisplayHeight - 50, W: DisplayWidth, H: 50})
}

func (h *MainScreen) DrawTitle() {
	DrawTexture(h.renderer, IconTrimuiTexture, sdl.Rect{X: 20, Y: 15, W: 40, H: 40})
	DrawText(h.renderer, h.GetTitle(), sdl.Point{X: 75, Y: 10}, ContentColor1, ContentFont1)
}

func (h *MainScreen) DrawTip(renderer *sdl.Renderer, texture *sdl.Texture, text string, position sdl.Point) int32 {
	width := int32(30)
	_, _, width, _, err := texture.Query()
	if err != nil {
		width = 30
	}
	DrawTexture(renderer, texture, sdl.Rect{X: position.X, Y: position.Y, W: width, H: 30})
	DrawTextCenter(renderer, text, sdl.Rect{X: position.X + width + 10, Y: position.Y, W: 30, H: 30}, ContentColor1, ContentFont2, true, false)
	return position.X + width + 10 + int32(TextWidth(ContentFont2, text)) + 10
}

func (h *MainScreen) DrawItemTips(item *menuItem, xPos int32) int32 {
	tipPos := sdl.Point{X: xPos, Y: DisplayHeight - 40}

	if !item.Disabled {

		if item.Type == "cmd" {
			tipPos.X = h.DrawTip(h.renderer, TipsATexture, "Run command", tipPos)
		}

		if item.Type == "menu" {
			tipPos.X = h.DrawTip(h.renderer, TipsATexture, "Open Folder", tipPos)
		}

		if item.Type == "toggle" {
			if item.Selected == 1 {
				tipPos.X = h.DrawTip(h.renderer, TipsATexture, "Disable", tipPos)
			} else {
				tipPos.X = h.DrawTip(h.renderer, TipsATexture, "Enable", tipPos)
			}
		}

		if item.Type == "select" {
			tipPos.X = h.DrawTip(h.renderer, TipsATexture, "Set option", tipPos)
		}
	}

	if len(h.lists) > 1 {
		tipPos.X = h.DrawTip(h.renderer, TipsBTexture, "Back", tipPos)
	} else {
		tipPos.X = h.DrawTip(h.renderer, TipsBTexture, "Exit", tipPos)
	}

	return tipPos.X
}

func (h *MainScreen) Draw() {
	h.InitMenu()
	tipPos := int32(20)
	_ = h.renderer.SetDrawColor(0, 0, 0, 255)
	_ = h.renderer.Clear()

	h.DrawThemeTextures()
	h.DrawTitle()
	if !h.initialized && h.showLineView {
		h.lineView.Draw()
	} else if h.InError() {
		h.errorTextView.Draw()
		tipPos += h.DrawTip(h.renderer, TipsBTexture, "Exit", sdl.Point{X: tipPos, Y: DisplayHeight - 40})
	} else {
		if h.InShell() {
			if h.showDialogView && h.IsExecuting() {
				h.dialogView.Draw()
				tipPos += h.DrawTip(h.renderer, TipsATexture, "OK", sdl.Point{X: tipPos, Y: DisplayHeight - 40})
				tipPos += h.DrawTip(h.renderer, TipsBTexture, "Cancel", sdl.Point{X: tipPos, Y: DisplayHeight - 40})
			} else {
				if h.showLineView {
					h.lineView.Draw()
				} else {
					h.shellView.Draw()
				}
				if h.IsExecuting() {
					tipPos += h.DrawTip(h.renderer, TipsBTexture, "Cancel", sdl.Point{X: tipPos, Y: DisplayHeight - 40})
				} else {
					tipPos += h.DrawTip(h.renderer, TipsBTexture, "Back", sdl.Point{X: tipPos, Y: DisplayHeight - 40})
				}
			}
		} else {
			h.GetList().Draw()
			tipPos += h.DrawItemTips(h.GetList().SelectedItem(), tipPos)
		}

		if len(h.shellView.GetText()) > 0 && !h.showShellView && !h.IsExecuting() {
			tipPos += h.DrawTip(h.renderer, TipsSelectTexture, "Show output", sdl.Point{X: tipPos, Y: DisplayHeight - 40})
		}
	}
	h.renderer.Present()
}
