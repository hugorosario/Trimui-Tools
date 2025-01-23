package app

import (
	"encoding/json"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

type menuItem struct {
	Label              string       `json:"label"`                 // label of the item
	Type               string       `json:"type"`                  // "toggle", "select", "menu", "cmd"
	Icon               string       `json:"icon,omitempty"`        // icon file path
	Description        string       `json:"description,omitempty"` // summary description of the item
	Load               string       `json:"load,omitempty"`        // command to execute when states needs to be loaded
	Execute            string       `json:"execute,omitempty"`     // command to execute
	Output             string       `json:"output,omitempty"`      // "line", "full", "none"
	SelectItems        []string     `json:"items,omitempty"`       // items for select type
	Selected           int          `json:"selected,omitempty"`    // index of the selected item
	MenuItems          []*menuItem  `json:"menu,omitempty"`        // sub menu items
	Disabled           bool         `json:"disable,omitempty"`     // if the item is disabled
	IconTexture        *sdl.Texture `json:"-"`                     // texture for the icon
	DisplayLabel       string       `json:"-"`                     // parsed label with tags replaced
	DisplayDescription string       `json:"-"`                     // parsed description with tags replaced
}

func (h *MainScreen) RenderListItem(renderer *sdl.Renderer, item *menuItem, itemRect sdl.Rect, selected bool) {
	if selected {
		DrawTexture(renderer, ListItemSelTexture, itemRect)
	} else {
		DrawTexture(renderer, ListItemTexture, itemRect)
	}

	labelHeight := TextHeight(ContentFont1, item.DisplayLabel)

	textPos := sdl.Point{X: itemRect.X + 30, Y: itemRect.Y + 35}
	if item.DisplayDescription != "" {
		textPos.Y = itemRect.Y + 20
	}

	if (item.Icon != "") && (item.IconTexture == nil) {
		item.IconTexture, _ = LoadTexture(renderer, item.Icon)
	}
	if (item.Type == "menu") && (item.IconTexture == nil) {
		item.IconTexture = FolderIconTexture
	}

	if item.IconTexture != nil {
		DrawTexture(renderer, item.IconTexture, sdl.Rect{X: itemRect.X + 20, Y: itemRect.Y + 20, W: 80, H: 80})
		textPos.X = itemRect.X + 110
	}

	DrawText(renderer, item.DisplayLabel, textPos, ContentColor1, ContentFont1)

	if item.DisplayDescription != "" {
		DrawText(renderer, item.DisplayDescription, sdl.Point{X: textPos.X, Y: textPos.Y + int32(labelHeight) - 5}, ContentColor2, ContentFont6)
	}

	switch item.Type {
	case "select":
		if selected {
			DrawTexture(renderer, OptionBgTexture, sdl.Rect{X: itemRect.X + itemRect.W - 490, Y: itemRect.Y + 30, W: 472, H: 60})
			if item.Selected == 0 {
				DrawTexture(renderer, LeftArrowDisabledTexture, sdl.Rect{X: itemRect.X + itemRect.W - 490, Y: itemRect.Y + 30, W: 60, H: 60})
			} else {
				DrawTexture(renderer, LeftArrowEnabledTexture, sdl.Rect{X: itemRect.X + itemRect.W - 490, Y: itemRect.Y + 30, W: 60, H: 60})
			}
			if item.Selected == len(item.SelectItems)-1 {
				DrawTexture(renderer, RightArrowDisabledTexture, sdl.Rect{X: itemRect.X + itemRect.W - 80, Y: itemRect.Y + 30, W: 60, H: 60})
			} else {
				DrawTexture(renderer, RightArrowEnabledTexture, sdl.Rect{X: itemRect.X + itemRect.W - 80, Y: itemRect.Y + 30, W: 60, H: 60})
			}
		} else {
			DrawTexture(renderer, LeftArrowDisabledTexture, sdl.Rect{X: itemRect.X + itemRect.W - 490, Y: itemRect.Y + 30, W: 60, H: 60})
			DrawTexture(renderer, RightArrowDisabledTexture, sdl.Rect{X: itemRect.X + itemRect.W - 80, Y: itemRect.Y + 30, W: 60, H: 60})
		}

		fromX := (itemRect.X + itemRect.W - 490 + 60)
		toX := (itemRect.X + itemRect.W - 80)
		DrawTextCenter(renderer, item.SelectItems[item.Selected], sdl.Rect{X: fromX, Y: itemRect.Y + 30, W: toX - fromX, H: 60}, ContentColor1, ContentFont1, true, true)
	case "toggle":
		if item.Selected == 1 {
			DrawTexture(renderer, SwitchOnTexture, sdl.Rect{X: itemRect.X + itemRect.W - 120, Y: itemRect.Y + 40, W: 90, H: 40})
		} else {
			DrawTexture(renderer, SwitchOffTexture, sdl.Rect{X: itemRect.X + itemRect.W - 120, Y: itemRect.Y + 40, W: 90, H: 40})
		}
	}

	if item.Disabled {
		//draw a semi-transparent black rectangle over the item
		renderer.SetDrawColor(0, 0, 0, 128)
		renderer.FillRect(&itemRect)
		renderer.SetDrawColor(0, 0, 0, 0)
	}
}

func (h *MainScreen) OpenSubmenu(item *menuItem) {
	if len(item.MenuItems) == 0 {
		return
	}
	//create a new list, add it to the lists slice and set it as the current list
	submenuList := NewList(
		h.renderer,
		int(visibleItems),
		sdl.Point{X: 20, Y: 72},
		func(index int, item Item[*menuItem], selected bool) {
			itemRect := sdl.Rect{X: 20, Y: 72 + (itemHeight+itemSpacing)*int32(index), W: 984, H: itemHeight}
			h.RenderListItem(h.renderer, item.Value, itemRect, selected)
		},
	)
	h.lists = append(h.lists, OptionsList{
		Title: item.DisplayLabel,
		List:  submenuList,
	})
	subitems := make([]Item[*menuItem], len(item.MenuItems))
	for i, item := range item.MenuItems {
		subitems[i] = Item[*menuItem]{
			Label: item.DisplayLabel,
			Value: item,
		}
	}
	h.title = h.lists[len(h.lists)-1].Title
	h.GetList().SetItems(subitems)
}

func (h *MainScreen) loadItems() ([]Item[*menuItem], error) {
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

	items := make([]Item[*menuItem], len(data))
	for i, item := range data {
		items[i] = Item[*menuItem]{
			Label: item.DisplayLabel,
			Value: &item,
		}
	}

	//run the load commands for all menu items
	for _, item := range items {
		h.runLoadCommands(item.Value)
	}
	return items, nil
}

func (h *MainScreen) DoItemAction(item *menuItem) {
	if item.Disabled {
		return
	}
	switch item.Type {
	case "menu":
		h.runLoadCommands(item)
		h.OpenSubmenu(item)
	case "cmd":
		h.RunCommandCaptured(item)
	case "toggle":
		item.Selected = 1 - item.Selected
		h.RunCommandCaptured(item)
	case "select":
		h.RunCommandCaptured(item)
	}
}
