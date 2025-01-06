package main

import (
	_ "embed"
	"log"
	"os"
	"runtime/debug"

	"github.com/hugorosario/trimuitools/gui"
	"github.com/hugorosario/trimuitools/input"
	"github.com/hugorosario/trimuitools/screens"

	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Unhandled error: %v\n", r)
			log.Println("Stack trace:")
			debug.PrintStack()
			os.Exit(-1)
		}
	}()

	if err := gui.InitSDL(); err != nil {
		panic(err)
	}

	if err := gui.InitTTF(); err != nil {
		panic(err)
	}

	window, err := sdl.CreateWindow("TrimuiTools", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, gui.ScreenWidth, gui.ScreenHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = window.Destroy()
	}()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = renderer.Destroy()
	}()

	if err := gui.InitTheme(renderer); err != nil {
		panic(err)
	}

	homeScreen, err := screens.NewHomeScreen(renderer)
	if err != nil {
		panic(err)
	}

	screensMap := map[string]func(){
		"home_screen": homeScreen.Draw,
	}

	inputHandlers := map[string]func(input.InputEvent){
		"home_screen": homeScreen.HandleInput,
	}

	input.StartListening()

	running := true
	for running {
		frameStart := sdl.GetTicks64()

		for {
			event := sdl.PollEvent()
			if event == nil {
				break
			}

			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
			}
		}

		select {
		case inputEvent := <-input.InputChannel:
			if handler, ok := inputHandlers[gui.CurrentScreen]; ok {
				handler(inputEvent)
			}
		default:
			// No event received
		}

		if drawFunc, ok := screensMap[gui.CurrentScreen]; ok {
			drawFunc()
		}

		//30 fps cap
		frameTime := sdl.GetTicks64() - frameStart
		if frameTime < 1000/30 {
			sdl.Delay(uint32((1000 / 30) - frameTime))
		}
	}
}
