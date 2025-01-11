package gui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func LoadThemeTexture(renderer *sdl.Renderer, asset string) (*sdl.Texture, error) {
	imagePath := fmt.Sprintf("%s%s", themeBasePath, asset)
	if !pathExists(imagePath) {
		imagePath = fmt.Sprintf("%s%s", DEFAULT_THEME_PATH, asset)
	}
	if pathExists(imagePath) {
		return LoadTexture(renderer, imagePath)
	}
	return LoadTexture(renderer, "./empty.png")
}

func LoadTexture(renderer *sdl.Renderer, imagePath string) (*sdl.Texture, error) {
	imgSurface, err := img.Load(imagePath)
	if err != nil {
		return nil, fmt.Errorf("error loading image: %w", err)
	}
	defer imgSurface.Free()

	texture, err := renderer.CreateTextureFromSurface(imgSurface)
	if err != nil {
		return nil, fmt.Errorf("error creating texture: %w", err)
	}
	return texture, nil
}

// DrawText is a function that draws text on the screen based on the provided position, color, and font.
func DrawText(renderer *sdl.Renderer, text string, position sdl.Point, color sdl.Color, font *ttf.Font) {
	// Render the text to a surface
	textSurface, err := RenderText(text, color, font)
	if err != nil {
		fmt.Printf("DrawText.Draw(%s): %v\n", text, err)
		return
	}
	defer textSurface.Free()

	// Create a texture from the surface
	textTexture, err := renderer.CreateTextureFromSurface(textSurface)
	if err != nil {
		fmt.Printf("Error creating texture: %v\n", err)
		return
	}
	defer func() {
		_ = textTexture.Destroy()
	}()

	// Set the destination rectangle for the texture
	destinationRect := sdl.Rect{
		X: position.X,
		Y: position.Y,
		W: textSurface.W,
		H: textSurface.H,
	}

	// Copy the texture to the renderer
	_ = renderer.Copy(textTexture, nil, &destinationRect)
}

// DrawTextCenter is a function that draws text on the screen center in the provided rect, color, and font.
func DrawTextCenter(renderer *sdl.Renderer, text string, rect sdl.Rect, color sdl.Color, font *ttf.Font, verticalAlign bool, horizontalAlign bool) {
	// Render the text to a surface
	textSurface, err := RenderText(text, color, font)
	if err != nil {
		fmt.Printf("DrawTextCenter.Draw(%s): %v\n", text, err)
		return
	}
	defer textSurface.Free()

	// Create a texture from the surface
	textTexture, err := renderer.CreateTextureFromSurface(textSurface)
	if err != nil {
		fmt.Printf("Error creating texture: %v\n", err)
		return
	}
	defer func() {
		_ = textTexture.Destroy()
	}()

	// render at the horizontal center of the rect
	destinationRect := sdl.Rect{
		X: rect.X,
		Y: rect.Y,
		W: textSurface.W,
		H: textSurface.H,
	}

	if horizontalAlign {
		destinationRect.X = rect.X + (rect.W-textSurface.W)/2
	}

	if verticalAlign {
		destinationRect.Y = rect.Y + (rect.H-textSurface.H)/2
	}

	// Copy the texture to the renderer
	_ = renderer.Copy(textTexture, nil, &destinationRect)
}

// RenderText renders text to an SDL surface
func RenderText(text string, color sdl.Color, font *ttf.Font) (*sdl.Surface, error) {
	textSurface, err := font.RenderUTF8Blended(text, color)
	if err != nil {
		return nil, fmt.Errorf("error rendering text: %w", err)
	}
	return textSurface, nil
}

func DrawTexture(renderer *sdl.Renderer, texture *sdl.Texture, rect sdl.Rect) {
	_ = renderer.Copy(texture, nil, &rect)
}

// WrapText splits a long text into multiple lines based on the specified maximum width.
func WrapText(text string, charWidth int, maxWidth int) []string {
	var wrappedLines []string
	words := strings.Split(text, " ")
	if len(words) == 0 {
		return wrappedLines
	}

	var currentLine string
	for _, word := range words {
		// Measure the width of the current line with the new word
		lineWithWord := currentLine + " " + word
		width := charWidth * len(lineWithWord)
		if width > maxWidth {
			// If the line exceeds the max width, add the current line to wrappedLines
			// and start a new line with the current word
			if currentLine != "" {
				wrappedLines = append(wrappedLines, strings.TrimSpace(currentLine))
			}
			currentLine = word
		} else {
			// Otherwise, add the word to the current line
			currentLine = lineWithWord
		}
	}

	// Add the last line to wrappedLines
	if currentLine != "" {
		wrappedLines = append(wrappedLines, strings.TrimSpace(currentLine))
	}

	return wrappedLines
}

// TextWidth calculates the width of a string of text based on the provided font.
func TextWidth(font *ttf.Font, text string) int {
	surface, err := font.RenderUTF8Blended(text, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err != nil {
		return 0
	}
	defer surface.Free()

	return int(surface.W)
}

// TextHeight calculates the width of a string of text based on the provided font.
func TextHeight(font *ttf.Font, text string) int {
	surface, err := font.RenderUTF8Blended(text, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err != nil {
		return 0
	}
	defer surface.Free()

	return int(surface.H)
}

func readJsonFileProperty(filepath string, property string) (string, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	var data map[string]interface{}
	err = json.Unmarshal(file, &data)
	if err != nil {
		return "", err
	}

	keys := strings.Split(property, ".")
	var value interface{} = data

	for _, key := range keys {
		if m, ok := value.(map[string]interface{}); ok {
			value = m[key]
		} else {
			return "", fmt.Errorf("property %s not found in file %s", property, filepath)
		}
	}

	if strValue, ok := value.(string); ok {
		return strValue, nil
	}

	return "", fmt.Errorf("property %s not found in file %s", property, filepath)
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func loadJson(path string) (map[string]interface{}, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(file, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func HexToColor(cfontcolor1 string) sdl.Color {
	var a, r, g, b uint8
	fmt.Sscanf(cfontcolor1, "#%02x%02x%02x%02x", &a, &r, &g, &b)
	return sdl.Color{R: r, G: g, B: b, A: a}
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
