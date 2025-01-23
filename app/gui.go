package app

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	TERMINAL_FONT        = "./terminal.ttf"
	DEFAULT_THEME_PATH   = "/usr/trimui/res"
	THEME_BACKGROUND     = "/skin/bg.png"
	TIPS_BAR_BACKGROUND  = "/skin/tips-bar-bg.png"
	TITLE_BAR_BACKGROUND = "/skin/title-bg.png"
	TITLE_ICON           = "/skin/icon-trimui.png"
	LIST_ITEM_BACKGROUND = "/skin/list-item-1line-sort-bg-n.png"
	LIST_ITEM_SELECTED   = "/skin/list-item-1line-sort-bg-f.png"
	SWITCH_ON            = "/skin/sw-on.png"
	SWITCH_OFF           = "/skin/sw-off.png"
	LEFT_ARROW_ENABLED   = "/skin/ic-left-arrow-a.png"
	LEFT_ARROW_DISABLED  = "/skin/ic-left-arrow-n.png"
	RIGHT_ARROW_ENABLED  = "/skin/ic-right-arrow-a.png"
	RIGHT_ARROW_DISABLED = "/skin/ic-right-arrow-n.png"
	FOLDER_ICON          = "/skin/ic-folder.png"
	FILE_ICON            = "/skin/ic-file.png"
	OPTION_BG            = "/skin/bg-setting-textfield.png"
	TIPS_A               = "/skin/tips-A.png"
	TIPS_B               = "/skin/tips-B.png"
	TIPS_MENU            = "/skin/tips-MENU.png"
	TIPS_SELECT          = "/skin/tips-SELECT.png"
)

var (
	DisplayWidth  = int32(1024)
	DisplayHeight = int32(768)

	TerminalFont *ttf.Font
	ContentFont1 *ttf.Font
	ContentFont2 *ttf.Font
	ContentFont6 *ttf.Font

	BackgroundTexture         *sdl.Texture
	TipsBarTexture            *sdl.Texture
	TitleBarTexture           *sdl.Texture
	IconTrimuiTexture         *sdl.Texture
	ListItemTexture           *sdl.Texture
	ListItemSelTexture        *sdl.Texture
	SwitchOnTexture           *sdl.Texture
	SwitchOffTexture          *sdl.Texture
	LeftArrowEnabledTexture   *sdl.Texture
	LeftArrowDisabledTexture  *sdl.Texture
	RightArrowEnabledTexture  *sdl.Texture
	RightArrowDisabledTexture *sdl.Texture
	FolderIconTexture         *sdl.Texture
	FileIconTexture           *sdl.Texture
	OptionBgTexture           *sdl.Texture
	TipsATexture              *sdl.Texture
	TipsBTexture              *sdl.Texture
	TipsMenuTexture           *sdl.Texture
	TipsSelectTexture         *sdl.Texture

	ContentColor1 = HexToColor("#FFFFFFFF")
	ContentColor2 = HexToColor("#80FFFFFF")
)

var (
	themeBasePath = ""
	themeFont     = "/regular.ttf"
	window        *sdl.Window
	renderer      *sdl.Renderer
	screen        *MainScreen
	isRunning     bool
)

func InitGUI() error {
	err := initSDL()
	if err != nil {
		return err
	}

	err = initTTF()
	if err != nil {
		return err
	}

	window, err = sdl.CreateWindow("TrimuiTools", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, DisplayWidth, DisplayHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		return fmt.Errorf("error creating window: %w", err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return fmt.Errorf("error creating renderer: %w", err)
	}
	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	err = initTheme(renderer)
	if err != nil {
		return err
	}

	screen, err = NewMainScreen(renderer)
	if err != nil {
		return err
	}

	return nil
}

func Show() {
	go StartInputListener()
	isRunning = true
	for isRunning {
		frameStart := sdl.GetTicks64()

		for {
			event := sdl.PollEvent()
			if event == nil {
				break
			}

			switch event.(type) {
			case *sdl.QuitEvent:
				Close()
			}
		}

		select {
		case inputEvent := <-InputChannel:
			screen.HandleInput(inputEvent)
		default:
		}

		screen.Draw()

		//30 fps cap
		frameTime := sdl.GetTicks64() - frameStart
		if frameTime < 1000/30 {
			sdl.Delay(uint32((1000 / 30) - frameTime))
		}
	}
}

func Close() {
	isRunning = false
}

func FreeGUI() {

	//free textures
	if BackgroundTexture != nil {
		_ = BackgroundTexture.Destroy()
	}

	if TipsBarTexture != nil {
		_ = TipsBarTexture.Destroy()
	}

	if TitleBarTexture != nil {
		_ = TitleBarTexture.Destroy()
	}

	if IconTrimuiTexture != nil {
		_ = IconTrimuiTexture.Destroy()
	}

	if ListItemTexture != nil {
		_ = ListItemTexture.Destroy()
	}

	if ListItemSelTexture != nil {
		_ = ListItemSelTexture.Destroy()
	}

	if SwitchOnTexture != nil {
		_ = SwitchOnTexture.Destroy()
	}

	if SwitchOffTexture != nil {
		_ = SwitchOffTexture.Destroy()
	}

	if LeftArrowEnabledTexture != nil {
		_ = LeftArrowEnabledTexture.Destroy()
	}

	if LeftArrowDisabledTexture != nil {
		_ = LeftArrowDisabledTexture.Destroy()
	}

	if RightArrowEnabledTexture != nil {
		_ = RightArrowEnabledTexture.Destroy()
	}

	if RightArrowDisabledTexture != nil {
		_ = RightArrowDisabledTexture.Destroy()
	}

	if FolderIconTexture != nil {
		_ = FolderIconTexture.Destroy()
	}

	if FileIconTexture != nil {
		_ = FileIconTexture.Destroy()
	}

	if OptionBgTexture != nil {
		_ = OptionBgTexture.Destroy()
	}

	if TipsATexture != nil {
		_ = TipsATexture.Destroy()
	}

	if TipsBTexture != nil {
		_ = TipsBTexture.Destroy()
	}

	if TipsMenuTexture != nil {
		_ = TipsMenuTexture.Destroy()
	}

	if TipsSelectTexture != nil {
		_ = TipsSelectTexture.Destroy()
	}

	//free fonts

	if TerminalFont != nil {
		TerminalFont.Close()
	}

	if ContentFont1 != nil {
		ContentFont1.Close()
	}

	if ContentFont2 != nil {
		ContentFont2.Close()
	}

	if ContentFont6 != nil {
		ContentFont6.Close()
	}

	if renderer != nil {
		_ = renderer.Destroy()
	}

	if window != nil {
		_ = window.Destroy()
	}

	ttf.Quit()
	sdl.Quit()
}

func initTheme(renderer *sdl.Renderer) error {
	selectedTheme, err := readJsonFileProperty("/mnt/UDISK/system.json", "theme")
	if (err == nil) && (pathExists(selectedTheme)) {
		themeBasePath = selectedTheme
	} else {
		if pathExists("./theme") {
			themeBasePath = "./theme"
		} else {
			themeBasePath = DEFAULT_THEME_PATH
		}
	}

	if err := initFont(TERMINAL_FONT, &TerminalFont, 22); err != nil {
		panic(err)
	}

	if pathExists(themeBasePath + "/config.json") {
		configjson, err := loadJson(themeBasePath + "/config.json")
		if err != nil {
			panic(err)
		}

		configFont := configjson["font"].(string)
		if configFont != "" {
			themeFont = themeBasePath + "/" + configFont
		} else {
			themeFont = themeBasePath + "/regular.ttf"
		}

		cfontcolor1 := configjson["fontcolor"].(map[string]interface{})["content_color1"].(string)
		ContentColor1 = HexToColor(cfontcolor1)

		cfontcolor2 := configjson["fontcolor"].(map[string]interface{})["content_color2"].(string)
		ContentColor2 = HexToColor(cfontcolor2)

		cfont1 := configjson["fontsize"].(map[string]interface{})["content_font1"].(float64)
		if err := initThemeFont(&ContentFont1, int(cfont1)); err != nil {
			panic(err)
		}

		cfont2 := configjson["fontsize"].(map[string]interface{})["content_font2"].(float64)
		if err := initThemeFont(&ContentFont2, int(cfont2)); err != nil {
			panic(err)
		}

		cfont6 := configjson["fontsize"].(map[string]interface{})["content_font6"].(float64)
		if err := initThemeFont(&ContentFont6, int(cfont6)); err != nil {
			panic(err)
		}
	} else {
		themeFont = themeBasePath + "/regular.ttf"
		if err := initThemeFont(&ContentFont1, 36); err != nil {
			panic(err)
		}
		if err := initThemeFont(&ContentFont2, 24); err != nil {
			panic(err)
		}
		if err := initThemeFont(&ContentFont6, 28); err != nil {
			panic(err)
		}
	}

	BackgroundTexture, err = LoadThemeTexture(renderer, THEME_BACKGROUND)
	if err != nil {
		panic(err)
	}

	TipsBarTexture, err = LoadThemeTexture(renderer, TIPS_BAR_BACKGROUND)
	if err != nil {
		panic(err)
	}

	TitleBarTexture, err = LoadThemeTexture(renderer, TITLE_BAR_BACKGROUND)
	if err != nil {
		panic(err)
	}

	IconTrimuiTexture, err = LoadThemeTexture(renderer, TITLE_ICON)
	if err != nil {
		panic(err)
	}

	ListItemTexture, err = LoadThemeTexture(renderer, LIST_ITEM_BACKGROUND)
	if err != nil {
		panic(err)
	}

	ListItemSelTexture, err = LoadThemeTexture(renderer, LIST_ITEM_SELECTED)
	if err != nil {
		panic(err)
	}

	SwitchOnTexture, err = LoadThemeTexture(renderer, SWITCH_ON)
	if err != nil {
		panic(err)
	}

	SwitchOffTexture, err = LoadThemeTexture(renderer, SWITCH_OFF)
	if err != nil {
		panic(err)
	}

	LeftArrowEnabledTexture, err = LoadThemeTexture(renderer, LEFT_ARROW_ENABLED)
	if err != nil {
		panic(err)
	}

	LeftArrowDisabledTexture, err = LoadThemeTexture(renderer, LEFT_ARROW_DISABLED)
	if err != nil {
		panic(err)
	}

	RightArrowEnabledTexture, err = LoadThemeTexture(renderer, RIGHT_ARROW_ENABLED)
	if err != nil {
		panic(err)
	}

	RightArrowDisabledTexture, err = LoadThemeTexture(renderer, RIGHT_ARROW_DISABLED)
	if err != nil {
		panic(err)
	}

	FolderIconTexture, err = LoadThemeTexture(renderer, FOLDER_ICON)
	if err != nil {
		panic(err)
	}

	FileIconTexture, err = LoadThemeTexture(renderer, FILE_ICON)
	if err != nil {
		panic(err)
	}

	OptionBgTexture, err = LoadThemeTexture(renderer, OPTION_BG)
	if err != nil {
		panic(err)
	}

	TipsATexture, err = LoadThemeTexture(renderer, TIPS_A)
	if err != nil {
		panic(err)
	}

	TipsBTexture, err = LoadThemeTexture(renderer, TIPS_B)
	if err != nil {
		panic(err)
	}

	TipsMenuTexture, err = LoadThemeTexture(renderer, TIPS_MENU)
	if err != nil {
		panic(err)
	}

	TipsSelectTexture, err = LoadThemeTexture(renderer, TIPS_SELECT)
	if err != nil {
		panic(err)
	}

	return nil
}

func initSDL() error {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO | sdl.INIT_JOYSTICK | sdl.INIT_GAMECONTROLLER); err != nil {
		return fmt.Errorf("error initializing SDL: %w", err)
	}
	return nil
}

func initTTF() error {
	if err := ttf.Init(); err != nil {
		return fmt.Errorf("error initializing SDL_ttf: %w", err)
	}
	return nil
}

func initFont(path string, font **ttf.Font, size int) error {
	f, err := ttf.OpenFont(path, size)
	if err != nil {
		return fmt.Errorf("error loading font: %w", err)
	}
	*font = f
	return nil
}

func initThemeFont(font **ttf.Font, size int) error {
	return initFont(themeFont, font, size)
}
