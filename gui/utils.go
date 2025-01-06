package gui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hugorosario/trimuitools/output"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func LoadThemeTexture(renderer *sdl.Renderer, asset string) (*sdl.Texture, error) {
	imagePath := fmt.Sprintf("%s%s", ThemeBasePath, asset)
	if !pathExists(imagePath) {
		imagePath = fmt.Sprintf("%s%s", DEFAULT_THEME_PATH, asset)
	}
	if pathExists(imagePath) {
		return LoadTexture(renderer, imagePath)
	}
	return nil, fmt.Errorf("image not found: %s", asset)
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

// LoadFont loads a font from RWops and returns the font object
func LoadFont(rwops *sdl.RWops, size int) (*ttf.Font, error) {
	font, err := ttf.OpenFontRW(rwops, 1, size)
	if err != nil {
		return nil, fmt.Errorf("error loading font: %w", err)
	}
	return font, nil
}

// DrawText is a function that draws text on the screen based on the provided position, color, and font.
func DrawText(renderer *sdl.Renderer, text string, position sdl.Point, color sdl.Color, font *ttf.Font) {
	// Render the text to a surface
	textSurface, err := RenderText(text, color, font)
	if err != nil {
		output.Printf("Error rendering text: %v\n", err)
		return
	}
	defer textSurface.Free()

	// Create a texture from the surface
	textTexture, err := renderer.CreateTextureFromSurface(textSurface)
	if err != nil {
		output.Printf("Error creating texture: %v\n", err)
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

// RenderText renders text to an SDL surface
func RenderText(text string, color sdl.Color, font *ttf.Font) (*sdl.Surface, error) {
	textSurface, err := font.RenderUTF8Blended(text, color)
	if err != nil {
		return nil, fmt.Errorf("error rendering text: %w", err)
	}
	return textSurface, nil
}

func RenderTexture(renderer *sdl.Renderer, imagePath string, startQuadrant, endQuadrant string) {
	// Load the texture image

	textureSurface, err := sdl.LoadBMP(imagePath)
	if err != nil {
		output.Printf("Error loading texture image: %v\n", err)
		return
	}
	defer textureSurface.Free()

	textureTexture, err := renderer.CreateTextureFromSurface(textureSurface)
	if err != nil {
		output.Printf("Error creating texture from image: %v\n", err)
		return
	}
	defer func() {
		_ = textureTexture.Destroy()
	}()

	// Get screen width and height
	halfWidth, halfHeight := ScreenWidth/2, ScreenHeight/2

	// Define rectangles for each quadrant
	quadrants := map[string]sdl.Rect{
		"Q1": {X: halfWidth, Y: 0, W: halfWidth, H: halfHeight},          // Q1
		"Q2": {X: 0, Y: 0, W: halfWidth, H: halfHeight},                  // Q2
		"Q3": {X: 0, Y: halfHeight, W: halfWidth, H: halfHeight},         // Q3
		"Q4": {X: halfWidth, Y: halfHeight, W: halfWidth, H: halfHeight}, // Q4
	}

	// Check if the quadrants are valid
	startRect, startOk := quadrants[startQuadrant]
	endRect, endOk := quadrants[endQuadrant]

	if !startOk || !endOk {
		output.Printf("Unknown quadrant(s): %s, %s\n", startQuadrant, endQuadrant)
		return
	}

	// Calculate the rectangle covering the area between the quadrants
	dstRect := sdl.Rect{
		X: min(startRect.X, endRect.X),
		Y: min(startRect.Y, endRect.Y),
		W: max(startRect.X+startRect.W, endRect.X+endRect.W) - min(startRect.X, endRect.X),
		H: max(startRect.Y+startRect.H, endRect.Y+endRect.H) - min(startRect.Y, endRect.Y),
	}

	// Get the dimensions of the texture
	textureWidth, textureHeight := textureSurface.W, textureSurface.H

	// Calculate the source rectangle of the texture
	srcRect := sdl.Rect{
		X: 0,
		Y: 0,
		W: textureWidth,
		H: textureHeight,
	}

	// Render the texture adjusted to the area between the quadrants
	_ = renderer.Copy(textureTexture, &srcRect, &dstRect)
}

func RenderImage(renderer *sdl.Renderer, imagePath string) {
	textureSurface, err := img.Load(imagePath)
	if err != nil {
		output.Printf("Error loading texture image: %v\n", err)
		return
	}
	defer textureSurface.Free()

	textureTexture, err := renderer.CreateTextureFromSurface(textureSurface)
	if err != nil {
		output.Printf("Error creating texture from image: %v\n", err)
		return
	}
	defer func() {
		_ = textureTexture.Destroy()
	}()

	halfHeight := ScreenHeight / 2

	textureWidth, textureHeight := textureSurface.W, textureSurface.H
	imgWidth, imgHeight := ScreenWidth/5, ScreenHeight/5
	imgProportion := float64(imgWidth) / float64(imgHeight)
	imgWidthProportional := int32(float64(imgWidth) * imgProportion)

	dstRect := sdl.Rect{
		X: 890 - imgWidthProportional/2,
		Y: halfHeight - imgHeight/2,
		W: imgWidthProportional,
		H: imgHeight,
	}

	srcRect := sdl.Rect{
		X: 0,
		Y: 0,
		W: textureWidth,
		H: textureHeight,
	}

	_ = renderer.Copy(textureTexture, &srcRect, &dstRect)
}

func RenderImageAdjusted(renderer *sdl.Renderer, imagePath string, rect sdl.Rect) {
	// Load the texture image
	textureSurface, err := img.Load(imagePath)

	if err != nil {
		output.Printf("Error loading texture image: %v\n", err)
		return
	}
	defer textureSurface.Free()

	textureTexture, err := renderer.CreateTextureFromSurface(textureSurface)
	if err != nil {
		output.Printf("Error creating texture from image: %v\n", err)
		return
	}
	defer func() {
		_ = textureTexture.Destroy()
	}()

	// Draw the texture at the specified position and size
	_ = renderer.Copy(textureTexture, nil, &rect)
}

func RenderTextureAdjusted(renderer *sdl.Renderer, texture *sdl.Texture, rect sdl.Rect) {
	_ = renderer.Copy(texture, nil, &rect)
}

// WrapText splits a long text into multiple lines based on the specified maximum width.
func WrapText(text string, font *ttf.Font, maxWidth int) []string {
	words := strings.Fields(text)
	var lines []string
	var currentLine string

	for _, word := range words {
		lineWithWord := currentLine + word + " "
		lineWidth := textWidth(font, lineWithWord)

		if lineWidth > maxWidth {
			if len(currentLine) > 0 {
				lines = append(lines, strings.TrimSpace(currentLine))
			}
			currentLine = word + " "
		} else {
			currentLine = lineWithWord
		}
	}

	if len(currentLine) > 0 {
		lines = append(lines, strings.TrimSpace(currentLine))
	}

	return lines
}

// textWidth calculates the width of a string of text based on the provided font.
func textWidth(font *ttf.Font, text string) int {
	surface, err := font.RenderUTF8Blended(text, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err != nil {
		return 0
	}
	defer surface.Free()

	return int(surface.W)
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
