package gui

import (
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type InputEvent struct {
	KeyCode string
}

var InputChannel = make(chan InputEvent)

const InputReadDelay = 50

func StartInputListener() {
	controller := openController()
	defer func() {
		if controller != nil {
			controller.Close()
		}
	}()

	controllerMappings := map[sdl.GameControllerButton]string{
		sdl.CONTROLLER_BUTTON_DPAD_UP:    "UP",
		sdl.CONTROLLER_BUTTON_DPAD_DOWN:  "DOWN",
		sdl.CONTROLLER_BUTTON_DPAD_LEFT:  "LEFT",
		sdl.CONTROLLER_BUTTON_DPAD_RIGHT: "RIGHT",
		sdl.CONTROLLER_BUTTON_A:          "B",
		sdl.CONTROLLER_BUTTON_B:          "A",
		sdl.CONTROLLER_BUTTON_X:          "Y",
		sdl.CONTROLLER_BUTTON_Y:          "X",
		sdl.CONTROLLER_BUTTON_START:      "START",
		sdl.CONTROLLER_BUTTON_BACK:       "SELECT",
		sdl.CONTROLLER_BUTTON_GUIDE:      "MENU",
	}

	keyboardMappings := map[sdl.Scancode]string{
		sdl.SCANCODE_UP:     "UP",
		sdl.SCANCODE_DOWN:   "DOWN",
		sdl.SCANCODE_LEFT:   "LEFT",
		sdl.SCANCODE_RIGHT:  "RIGHT",
		sdl.SCANCODE_A:      "A",
		sdl.SCANCODE_B:      "B",
		sdl.SCANCODE_X:      "X",
		sdl.SCANCODE_Y:      "Y",
		sdl.SCANCODE_RETURN: "START",
		sdl.SCANCODE_SPACE:  "SELECT",
	}

	// State tracking for debounce and repeat
	previousButtonState := make(map[sdl.GameControllerButton]bool)
	firstPressTime := make(map[sdl.GameControllerButton]time.Time)
	lastEventTime := make(map[sdl.GameControllerButton]time.Time)

	previousKeyState := make(map[sdl.Scancode]bool)
	firstKeyPressTime := make(map[sdl.Scancode]time.Time)
	lastKeyEventTime := make(map[sdl.Scancode]time.Time)

	initialDelay := 500 * time.Millisecond

	for {
		sdl.PumpEvents()

		// Handle controller input
		for button, keyCode := range controllerMappings {
			if controller == nil {
				break
			}
			currentState := controller.Button(button) == sdl.PRESSED
			now := time.Now()

			if currentState {
				if !previousButtonState[button] {
					// Button just pressed
					firstPressTime[button] = now
					lastEventTime[button] = now
					InputChannel <- InputEvent{KeyCode: keyCode}
				} else {
					// Button held down
					if now.Sub(firstPressTime[button]) > initialDelay && now.Sub(lastEventTime[button]) > InputReadDelay {
						InputChannel <- InputEvent{KeyCode: keyCode}
						lastEventTime[button] = now
					}
				}
			}

			previousButtonState[button] = currentState
		}

		// Handle keyboard input
		keyboardState := sdl.GetKeyboardState()
		for scancode, keyCode := range keyboardMappings {
			currentState := keyboardState[scancode] == 1
			now := time.Now()

			if currentState {
				if !previousKeyState[scancode] {
					// Key just pressed
					firstKeyPressTime[scancode] = now
					lastKeyEventTime[scancode] = now
					InputChannel <- InputEvent{KeyCode: keyCode}
				} else {
					// Key held down
					if now.Sub(firstKeyPressTime[scancode]) > initialDelay && now.Sub(lastKeyEventTime[scancode]) > InputReadDelay {
						InputChannel <- InputEvent{KeyCode: keyCode}
						lastKeyEventTime[scancode] = now
					}
				}
			}

			previousKeyState[scancode] = currentState
		}

		sdl.Delay(InputReadDelay)
	}
}

func openController() *sdl.GameController {
	for i := 0; i < sdl.NumJoysticks(); i++ {
		if sdl.IsGameController(i) {
			controller := sdl.GameControllerOpen(i)
			if controller != nil {
				return controller
			}
		}
	}
	return nil
}
