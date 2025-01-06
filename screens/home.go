package screens

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hugorosario/trimuitools/components"
	"github.com/hugorosario/trimuitools/gui"
	"github.com/hugorosario/trimuitools/input"
	"github.com/hugorosario/trimuitools/output"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	itemHeight   = 100
	itemSpacing  = 8
	visibleItems = 6
)

type HomeScreen struct {
	renderer    *sdl.Renderer
	toolsList   *components.List[menuItem]
	textView    *components.TextView
	initialized bool
}

type menuItem struct {
	Label       string     `json:"label"`
	Type        string     `json:"type"` // "toggle", "select", "menu", "cmd"
	Icon        string     `json:"icon,omitempty"`
	Thumbnail   string     `json:"thumbnail,omitempty"`
	Execute     string     `json:"execute,omitempty"`
	SelectItems []string   `json:"items,omitempty"`
	Selected    string     `json:"selected,omitempty"`
	MenuItems   []menuItem `json:"menu,omitempty"`
}

func RenderItem(renderer *sdl.Renderer, item menuItem, itemRect sdl.Rect, selected bool) {
	color := gui.Colors.WHITE
	if selected {
		color = gui.Colors.PRIMARY
	}
	textSurface, err := gui.RenderText(item.Label, color, gui.ListFont)
	if err != nil {
		output.Printf("Error rendering text: %v\n", err)
		return
	}
	defer textSurface.Free()

	texture, err := renderer.CreateTextureFromSurface(textSurface)
	if err != nil {
		output.Printf("Error creating texture: %v\n", err)
		return
	}

	defer func() { _ = texture.Destroy() }()

	if selected {
		gui.RenderTextureAdjusted(renderer, gui.ListItemSelTexture, itemRect)
	} else {
		gui.RenderTextureAdjusted(renderer, gui.ListItemTexture, itemRect)
	}

	renderer.Copy(texture, nil, &sdl.Rect{X: itemRect.X + 25, Y: itemRect.Y + 30, W: textSurface.W, H: textSurface.H})

	switch item.Type {
	case "menu":
		RenderMenuItem(renderer, item, itemRect, selected)
	case "cmd":
		RenderCommandItem(renderer, item, itemRect, selected)
	case "select":
		RenderSelectItem(renderer, item, itemRect, selected)
	case "toggle":
		RenderToggleItem(renderer, item, itemRect, selected)
	}
}
func RenderToggleItem(renderer *sdl.Renderer, item menuItem, itemRect sdl.Rect, selected bool) {
	upperCaseSel := strings.ToUpper(item.Selected)
	textSurface, err := gui.RenderText(upperCaseSel, gui.Colors.WHITE, gui.ListFont)
	if err != nil {
		output.Printf("Error rendering text: %v\n", err)
		return
	}
	defer textSurface.Free()

	texture, err := renderer.CreateTextureFromSurface(textSurface)
	if err != nil {
		output.Printf("Error creating texture: %v\n", err)
		return
	}
	defer func() { _ = texture.Destroy() }()
	renderer.Copy(texture, nil, &sdl.Rect{X: itemRect.W - textSurface.W, Y: itemRect.Y + 30, W: textSurface.W, H: textSurface.H})
}

func RenderMenuItem(renderer *sdl.Renderer, item menuItem, itemRect sdl.Rect, selected bool) {
	textSurface, err := gui.RenderText("+", gui.Colors.WHITE, gui.ListFont)
	if err != nil {
		output.Printf("Error rendering text: %v\n", err)
		return
	}
	defer textSurface.Free()

	texture, err := renderer.CreateTextureFromSurface(textSurface)
	if err != nil {
		output.Printf("Error creating texture: %v\n", err)
		return
	}
	defer func() { _ = texture.Destroy() }()
	renderer.Copy(texture, nil, &sdl.Rect{X: itemRect.W - textSurface.W, Y: itemRect.Y + 30, W: textSurface.W, H: textSurface.H})
}

func RenderSelectItem(renderer *sdl.Renderer, item menuItem, itemRect sdl.Rect, selected bool) {
	var selItem = fmt.Sprintf("< %s >", item.Selected)
	textSurface, err := gui.RenderText(selItem, gui.Colors.WHITE, gui.ListFont)
	if err != nil {
		output.Printf("Error rendering text: %v\n", err)
		return
	}
	defer textSurface.Free()

	texture, err := renderer.CreateTextureFromSurface(textSurface)
	if err != nil {
		output.Printf("Error creating texture: %v\n", err)
		return
	}
	defer func() { _ = texture.Destroy() }()
	renderer.Copy(texture, nil, &sdl.Rect{X: itemRect.W - textSurface.W, Y: itemRect.Y + 30, W: textSurface.W, H: textSurface.H})
}

func RenderCommandItem(renderer *sdl.Renderer, item menuItem, itemRect sdl.Rect, selected bool) {
	//TODO render the execute icon
}

func NewHomeScreen(renderer *sdl.Renderer) (*HomeScreen, error) {
	var listPosition = sdl.Point{X: 20, Y: 72}
	return &HomeScreen{
		renderer: renderer,
		toolsList: components.NewList(
			renderer,
			int(visibleItems),
			listPosition,
			func(index int, item components.Item[menuItem], selected bool) {
				itemRect := sdl.Rect{X: listPosition.X, Y: listPosition.Y + (itemHeight+itemSpacing)*int32(index), W: 984, H: itemHeight}
				RenderItem(renderer, item.Value, itemRect, selected)
			},
		),
		textView: components.NewTextView(
			renderer,
			components.TextViewSize{Width: 50, Height: 18},
			sdl.Point{X: 20, Y: 80},
		),
	}, nil
}

func (h *HomeScreen) InitHome() {
	if h.initialized {
		return
	}
	items, err := loadItems()
	if err != nil {
		h.textView.AddText(fmt.Sprintf("Error loading items: %s", err))
	} else {
		h.toolsList.SetItems(items)
	}
	h.initialized = true
}

func (h *HomeScreen) HandleInput(event input.InputEvent) {
	switch event.KeyCode {
	case "DOWN":
		h.toolsList.ScrollDown()
	case "UP":
		h.toolsList.ScrollUp()
	case "B":
		os.Exit(0)
	case "LEFT":
		h.ChangeSelection(h.toolsList.SelectedItem(), "LEFT")
	case "RIGHT":
		h.ChangeSelection(h.toolsList.SelectedItem(), "RIGHT")
	case "A":
		if h.isNotInErrorMode() {
			h.DoItemAction(h.toolsList.SelectedItem())
		}
	}
}

func (h *HomeScreen) isNotInErrorMode() bool {
	return len(h.textView.GetText()) == 0
}

func (h *HomeScreen) ChangeSelection(item *menuItem, direction string) {
	switch item.Type {
	case "select":
		if direction == "RIGHT" {
			for i, selectItem := range item.SelectItems {
				if selectItem == item.Selected {
					if i < len(item.SelectItems)-1 {
						item.Selected = item.SelectItems[i+1]
					}
					break
				}
			}
		} else {
			for i, selectItem := range item.SelectItems {
				if selectItem == item.Selected {
					if i > 0 {
						item.Selected = item.SelectItems[i-1]
					}
					break
				}
			}
		}
	}
}

func (h *HomeScreen) DoItemAction(item *menuItem) {
	switch item.Type {
	case "menu":
		h.OpenSubmenu(item.MenuItems)
	case "cmd":
		//TODO execute command
	case "toggle":
		if strings.ToUpper(item.Selected) == "ON" {
			item.Selected = "OFF"
		} else {
			item.Selected = "ON"
		}
	}
}

func (h *HomeScreen) OpenSubmenu(systems []menuItem) {
	// if len(systems) == 0 {
	// 	return
	// }
	// config.CurrentScreen = "scraping_screen"
	// SetTargetSystems(systems)
	// SetScraping()
	// h.initialized = false
}

func (h *HomeScreen) updateLogo() {
	// selectedSystem := h.toolsList.SelectedValue()
	// logoPath := fmt.Sprintf("%s/%s.png", config.LogosBaseDir, selectedSystem.DirName)

	// if _, err := os.Stat(logoPath); err == nil {
	// 	uilib.RenderImage(h.renderer, logoPath)
	// }
}

func (h *HomeScreen) Draw() {
	h.InitHome()

	_ = h.renderer.SetDrawColor(0, 0, 0, 255)
	_ = h.renderer.Clear()

	gui.RenderTextureAdjusted(h.renderer, gui.BackgroundTexture, sdl.Rect{X: 0, Y: 0, W: gui.ScreenWidth, H: gui.ScreenHeight})
	gui.RenderTextureAdjusted(h.renderer, gui.TitleBarTexture, sdl.Rect{X: 0, Y: 0, W: gui.ScreenWidth, H: 70})
	gui.RenderTextureAdjusted(h.renderer, gui.TipsBarTexture, sdl.Rect{X: 0, Y: gui.ScreenHeight - 50, W: gui.ScreenWidth, H: 50})
	gui.RenderTextureAdjusted(h.renderer, gui.IconTrimuiTexture, sdl.Rect{X: 20, Y: 15, W: 40, H: 40})
	gui.DrawText(h.renderer, "Tools", sdl.Point{X: 75, Y: 10}, gui.Colors.WHITE, gui.HeaderFont)

	h.toolsList.Draw(gui.Colors.WHITE, gui.Colors.SECONDARY)

	if len(h.textView.GetText()) > 0 {
		h.textView.Draw(gui.Colors.WHITE)
	} else {
		h.updateLogo()
	}

	h.renderer.Present()
}

func loadItems() ([]components.Item[menuItem], error) {

	//load items from the menu.json file
	file, err := os.ReadFile("./menu.json")
	if err != nil {
		return nil, err
	}

	var data []menuItem
	err = json.Unmarshal(file, &data)
	if err != nil {
		return nil, err
	}

	items := make([]components.Item[menuItem], len(data))
	for i, item := range data {
		items[i] = components.Item[menuItem]{
			Label: item.Label,
			Value: item,
		}
	}
	return items, nil
}
