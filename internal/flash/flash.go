package flash

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	"github.com/markel149/midi-flash-app/internal/config"
)

// Controller manages the flash rectangle and its behavior
type Controller struct {
	rect      *canvas.Rectangle
	flashCh   chan struct{}
	config    *config.FlashConfig
}

// NewController creates a new flash controller
func NewController(rect *canvas.Rectangle, cfg *config.FlashConfig) *Controller {
	return &Controller{
		rect:      rect,
		flashCh:   make(chan struct{}, 1),
		config:    cfg,
	}
}

// Start begins listening for flash events
func (c *Controller) Start() {
	go func() {
		for range c.flashCh {
			for i := 0; i < c.config.Repetitions; i++ {
				c.flash()
			}
		}
	}()
}

// Trigger sends a flash event (non-blocking)
func (c *Controller) Trigger() {
	select {
	case c.flashCh <- struct{}{}:
	default:
	}
}

// flash performs a single flash cycle
func (c *Controller) flash() {
	fyne.Do(func() {
		c.rect.FillColor = c.config.Color
		c.rect.Refresh()
	})
	time.Sleep(time.Duration(c.config.TimeMs) * time.Millisecond)
	fyne.Do(func() {
		c.rect.FillColor = color.Black
		c.rect.Refresh()
	})
	time.Sleep(time.Duration(c.config.DelayMs) * time.Millisecond)
}

// UpdateConfig updates the flash configuration
func (c *Controller) UpdateConfig(cfg *config.FlashConfig) {
	c.config = cfg
}
