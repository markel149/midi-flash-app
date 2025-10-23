package config

import "image/color"

// FlashConfig holds the configuration for the flash behavior
type FlashConfig struct {
	TimeMs       int
	Repetitions  int
	DelayMs      int
	Color        color.RGBA
}

// NewFlashConfig creates a new FlashConfig with default values
func NewFlashConfig() *FlashConfig {
	return &FlashConfig{
		TimeMs:      100,
		Repetitions: 1,
		DelayMs:     50,
		Color:       color.RGBA{R: 255, G: 0, B: 0, A: 255}, // Red
	}
}

// GetColorByName returns a color.RGBA based on a color name
func GetColorByName(name string) color.RGBA {
	switch name {
	case "Rojo":
		return color.RGBA{R: 255, G: 0, B: 0, A: 255}
	case "Verde":
		return color.RGBA{R: 0, G: 255, B: 0, A: 255}
	case "Azul":
		return color.RGBA{R: 0, G: 0, B: 255, A: 255}
	case "Blanco":
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	case "Amarillo":
		return color.RGBA{R: 255, G: 255, B: 0, A: 255}
	case "Cian":
		return color.RGBA{R: 0, G: 255, B: 255, A: 255}
	case "Magenta":
		return color.RGBA{R: 255, G: 0, B: 255, A: 255}
	default:
		return color.RGBA{R: 255, G: 0, B: 0, A: 255} // Default to red
	}
}

// ColorNames returns the list of available color names
func ColorNames() []string {
	return []string{"Rojo", "Verde", "Azul", "Blanco", "Amarillo", "Cian", "Magenta"}
}
