package gui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

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

type menuItem struct {
	Label              string       `json:"label"`                 // label of the item
	Type               string       `json:"type"`                  // "toggle", "select", "menu", "cmd"
	Icon               string       `json:"icon,omitempty"`        // icon file path
	Description        string       `json:"description,omitempty"` // summary description of the item
	Load               string       `json:"load,omitempty"`        // command to execute when status needs to be loaded
	Execute            string       `json:"execute,omitempty"`     // command to execute
	Output             string       `json:"output,omitempty"`      // "line", "full", "none"
	SelectItems        []string     `json:"items,omitempty"`       // items for select type
	Selected           int          `json:"selected,omitempty"`    // index of the selected item
	MenuItems          []*menuItem  `json:"menu,omitempty"`        // sub menu items
	IconTexture        *sdl.Texture `json:"-"`                     // texture for the icon
	DisplayLabel       string       `json:"-"`                     // parsed label with tags replaced
	DisplayDescription string       `json:"-"`                     // parsed description with tags replaced
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
}

func (h *MainScreen) InitHome() {
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

func (h *MainScreen) DoItemAction(item *menuItem) {
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

func (h *MainScreen) RunCommandCaptured(item *menuItem) {
	if item.Execute == "" {
		return
	}
	h.showLineView = item.Output == "line"
	h.showShellView = item.Output == "full"
	h.showDialogView = false
	if h.showLineView || h.showShellView {
		h.title = item.DisplayLabel
	}

	cmd := item.Execute
	if (item.Type == "select") || (item.Type == "toggle") {
		cmd = fmt.Sprintf("%s %d", cmd, item.Selected)
	}
	h.shellView.Clear()
	h.lineView.SetText("")
	h.dialogView.SetText("")
	h.shellInput, h.shellCancel, _ = shell.RunCommandAsync(cmd, func(event shell.ShellEvent) {
		switch event.Type {
		case "start":
			h.shellRunnning = true
			h.shellView.AddText("Executing: " + event.Data)
		case "output":
			//if contains tags in between {{ and }}, ignore this output
			if strings.Contains(event.Data, "{{") && strings.Contains(event.Data, "}}") {
				return
			}
			data := strings.ReplaceAll(event.Data, "\\n", "\n")
			//example of confirm dialog: confirm:ok:cancel:Are you sure you want to continue
			//check if prefix is for showing a confirm dialog
			if strings.HasPrefix(data, "confirm:") {
				h.dialogView.SetText(strings.TrimPrefix(data, "confirm:"))
				h.title = item.DisplayLabel
				h.showDialogView = true
			} else if h.showLineView {
				h.lineView.SetText(data)
			}
			h.shellView.AddText(data)
		case "error":
			h.shellView.AddText(event.Data)
		case "exit":
			h.showDialogView = false
			h.shellView.AddText("Exit code: " + event.Data)
			exitCode, err := strconv.Atoi(event.Data)
			if err != nil {
				h.shellView.AddText("Error parsing exit code: " + err.Error())
				exitCode = 1
			}
			if exitCode != 0 {
				h.showLineView = false
				h.showShellView = true
			}
			if h.shellCancel != nil {
				h.shellCancel()
			}
			if (item.Type == "toggle") && (exitCode != 0) {
				item.Selected = 1 - item.Selected
			}

			h.runLoadCommands(item)
		case "end":
			h.shellRunnning = false
			h.title = h.lists[len(h.lists)-1].Title
			//run load commands for all menu items
			for _, item := range h.GetList().GetItems() {
				h.runLoadCommands(item.Value)
			}
		}
	})
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

	if len(h.lists) > 1 {
		tipPos.X = h.DrawTip(h.renderer, TipsBTexture, "Back", tipPos)
	} else {
		tipPos.X = h.DrawTip(h.renderer, TipsBTexture, "Exit", tipPos)
	}

	return tipPos.X
}

func (h *MainScreen) Draw() {
	h.InitHome()
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

func (h *MainScreen) runLoadCommands(item *menuItem) {
	item.DisplayLabel = item.Label
	item.DisplayDescription = item.Description
	labelTags := getTags(item.Label)
	descriptionTags := getTags(item.Description)

	if item.Load != "" {
		cmd := item.Load
		output, err := shell.RunCommandSync(cmd)
		if err == nil {
			//iterate all lines in the output
			for _, line := range strings.Split(output, "\n") {
				//find if the line contains any tag values in the format of {key=value}
				tags := getTags(line)
				for _, tag := range tags {
					//split the tag into key and value
					parts := strings.Split(tag, "=")
					if len(parts) != 2 {
						continue
					}
					key := parts[0]
					value := parts[1]

					if (key == "state") && (strings.Contains("toggle,select", item.Type)) {
						item.Selected, err = strconv.Atoi(value)
						if err != nil {
							continue
						}
					}

					//replace the tag in the label and description
					for i, labelTag := range labelTags {
						if strings.Contains(labelTag, key) {
							item.DisplayLabel = strings.ReplaceAll(item.DisplayLabel, "{{"+labelTag+"}}", value)
							labelTags = append(labelTags[:i], labelTags[i+1:]...)
						}
					}

					for i, descriptionTag := range descriptionTags {
						if strings.Contains(descriptionTag, key) {
							item.DisplayDescription = strings.ReplaceAll(item.DisplayDescription, "{{"+descriptionTag+"}}", value)
							descriptionTags = append(descriptionTags[:i], descriptionTags[i+1:]...)
						}
					}
				}
			}
		}
	}

	//replace any remaining tags with empty strings
	for _, labelTag := range labelTags {
		item.DisplayLabel = strings.ReplaceAll(item.DisplayLabel, "{{"+labelTag+"}}", "")
	}

	for _, descriptionTag := range descriptionTags {
		item.DisplayDescription = strings.ReplaceAll(item.DisplayDescription, "{{"+descriptionTag+"}}", "")
	}

	//recursively run the load commands for all menu items
	for _, subitem := range item.MenuItems {
		h.runLoadCommands(subitem)
	}
}

func getTags(s string) []string {
	//extract all tags that are demlimited by {{ and }}
	tags := make([]string, 0)
	for _, tag := range strings.Split(s, "{{") {
		if strings.Contains(tag, "}}") {
			tags = append(tags, strings.Split(tag, "}}")[0])
		}
	}
	return tags
}
