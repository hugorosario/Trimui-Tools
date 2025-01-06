package gui

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	DEFAULT_THEME_PATH   = "/usr/trimui/res"
	THEME_BACKGROUND     = "/skin/bg.png"
	TIPS_BAR_BACKGROUND  = "/skin/tips-bar-bg.png"
	TITLE_BAR_BACKGROUND = "/skin/title-bg.png"
	TITLE_ICON           = "/skin/icon-trimui.png"
	LIST_ITEM_BACKGROUND = "/skin/list-item-1line-sort-bg-n.png"
	LIST_ITEM_SELECTED   = "/skin/list-item-1line-sort-bg-f.png"
)

var (
	ThemeBasePath      = ""
	ThemeFont          = "/msyh.ttf"
	ScreenWidth        = int32(1024)
	ScreenHeight       = int32(768)
	CurrentScreen      string
	BodyFont           *ttf.Font
	HeaderFont         *ttf.Font
	ListFont           *ttf.Font
	LongTextFont       *ttf.Font
	BackgroundTexture  *sdl.Texture
	TipsBarTexture     *sdl.Texture
	TitleBarTexture    *sdl.Texture
	IconTrimuiTexture  *sdl.Texture
	ListItemTexture    *sdl.Texture
	ListItemSelTexture *sdl.Texture
	Colors             FontColors
)

type FontColors struct {
	WHITE     sdl.Color
	PRIMARY   sdl.Color
	SECONDARY sdl.Color
	BLACK     sdl.Color
}

func InitTheme(renderer *sdl.Renderer) error {
	CurrentScreen = "home_screen"
	BodyFont = nil
	HeaderFont = nil
	ListFont = nil
	LongTextFont = nil
	Colors = FontColors{
		WHITE:     sdl.Color{R: 255, G: 255, B: 255, A: 255},
		PRIMARY:   sdl.Color{R: 113, G: 255, B: 142, A: 255},
		SECONDARY: sdl.Color{R: 168, G: 48, B: 190, A: 255},
		BLACK:     sdl.Color{R: 0, G: 0, B: 0, A: 255},
	}
	selectedTheme, err := readJsonFileProperty("/mnt/UDISK/system.json", "theme")
	if err == nil {
		ThemeBasePath = selectedTheme
	} else {
		if pathExists("./theme") {
			ThemeBasePath = "./theme"
		} else {
			ThemeBasePath = DEFAULT_THEME_PATH
		}
	}
	configFont, err := readJsonFileProperty(ThemeBasePath+"/config.json", "font")
	if err == nil {
		ThemeFont = ThemeBasePath + "/" + configFont
	} else {
		ThemeFont = ThemeBasePath + "/msyh.ttf"
	}

	if err := InitFont(&BodyFont, 30); err != nil {
		panic(err)
	}

	if err := InitFont(&ListFont, 30); err != nil {
		panic(err)
	}

	if err := InitFont(&LongTextFont, 20); err != nil {
		panic(err)
	}

	if err := InitFont(&HeaderFont, 36); err != nil {
		panic(err)
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

	return nil
}

func InitSDL() error {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO | sdl.INIT_JOYSTICK | sdl.INIT_GAMECONTROLLER); err != nil {
		return fmt.Errorf("error initializing SDL: %w", err)
	}
	return nil
}

func InitTTF() error {
	if err := ttf.Init(); err != nil {
		return fmt.Errorf("error initializing SDL_ttf: %w", err)
	}
	return nil
}

func InitFont(font **ttf.Font, size int) error {
	f, err := ttf.OpenFont(ThemeFont, size)
	if err != nil {
		return fmt.Errorf("error loading font: %w", err)
	}
	*font = f
	return nil
}
