package ui

import (
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"gitlab.com/gomidi/midi/v2/drivers"

	"github.com/markel149/midi-flash-app/internal/config"
	"github.com/markel149/midi-flash-app/internal/flash"
	"github.com/markel149/midi-flash-app/internal/midi"
)

// Components holds all UI components
type Components struct {
	FlashRect      *canvas.Rectangle
	FlashContainer *fyne.Container
	MessageBox     *fyne.Container
	ScrollMessages *container.Scroll
	InPortsList    *widget.Entry
	OutPortsList   *widget.Entry
	FlashEntry     *widget.Entry
	RepetitionsEntry *widget.Entry
	DelayEntry     *widget.Entry
	ColorSelect    *widget.Select
	MIDISelect     *widget.Select
	Tabs           *container.AppTabs
	Config         *config.FlashConfig
	FlashController *flash.Controller
	MIDIManager    *midi.Manager
}

// NewComponents creates and initializes all UI components
func NewComponents(midiManager *midi.Manager) *Components {
	// Create flash rectangle
	flashRect := canvas.NewRectangle(color.Black)
	flashRect.SetMinSize(fyne.NewSize(100, 100))
	flashContainer := container.NewMax(flashRect)

	// Create message box
	msgBox := container.NewVBox()
	scrollMessages := container.NewVScroll(msgBox)
	scrollMessages.SetMinSize(fyne.NewSize(350, 100))

	// Create configuration entries
	inPortsList := widget.NewMultiLineEntry()
	inPortsList.SetPlaceHolder("Puertos MIDI detectados...")
	inPortsList.Disable()

	outPortsList := widget.NewMultiLineEntry()
	outPortsList.SetPlaceHolder("Puertos MIDI de salida...")
	outPortsList.Disable()

	// Create flash configuration
	cfg := config.NewFlashConfig()
	
	flashEntry := widget.NewEntry()
	flashEntry.SetText(strconv.Itoa(cfg.TimeMs))

	repetitionsEntry := widget.NewEntry()
	repetitionsEntry.SetText(strconv.Itoa(cfg.Repetitions))

	delayEntry := widget.NewEntry()
	delayEntry.SetText(strconv.Itoa(cfg.DelayMs))

	// Create flash controller
	flashController := flash.NewController(flashRect, cfg)
	flashController.Start()

	c := &Components{
		FlashRect:        flashRect,
		FlashContainer:   flashContainer,
		MessageBox:       msgBox,
		ScrollMessages:   scrollMessages,
		InPortsList:      inPortsList,
		OutPortsList:     outPortsList,
		FlashEntry:       flashEntry,
		RepetitionsEntry: repetitionsEntry,
		DelayEntry:       delayEntry,
		Config:           cfg,
		FlashController:  flashController,
		MIDIManager:      midiManager,
	}

	// Color selector
	c.ColorSelect = c.createColorSelect()

	// MIDI port selector
	c.MIDISelect = c.createMIDISelect()

	// Create tabs
	c.Tabs = c.createTabs()

	return c
}

// createColorSelect creates the color selection widget
func (c *Components) createColorSelect() *widget.Select {
	colorSelect := widget.NewSelect(config.ColorNames(), func(s string) {
		c.Config.Color = config.GetColorByName(s)
	})
	colorSelect.SetSelected("Rojo")
	return colorSelect
}

// createMIDISelect creates the MIDI port selection widget
func (c *Components) createMIDISelect() *widget.Select {
	ins, _ := c.MIDIManager.GetInputPorts()
	names := []string{}
	for _, in := range ins {
		names = append(names, in.String())
	}

	midiSelect := widget.NewSelect(names, func(selected string) {
		ins, _ := c.MIDIManager.GetInputPorts()
		for _, in := range ins {
			if in.String() == selected {
				c.startMIDI(in)
				break
			}
		}
	})
	return midiSelect
}

// startMIDI starts listening to a MIDI port
func (c *Components) startMIDI(in drivers.In) {
	err := c.MIDIManager.ListenToPort(
		in,
		func(message string) {
			c.addMessage(message)
		},
		func() {
			c.FlashController.Trigger()
		},
	)
	if err != nil {
		// Log error but don't crash
		println("Error starting MIDI:", err.Error())
	}
}

// addMessage adds a message to the message box
func (c *Components) addMessage(msg string) {
	fyne.Do(func() {
		msgLabel := widget.NewLabel(msg)
		c.MessageBox.Add(msgLabel)
		c.ScrollMessages.ScrollToBottom()
	})
}

// UpdateFlashConfig updates the flash configuration from UI entries
func (c *Components) UpdateFlashConfig() {
	if val, err := strconv.Atoi(c.FlashEntry.Text); err == nil && val > 0 {
		c.Config.TimeMs = val
	}
	if val, err := strconv.Atoi(c.RepetitionsEntry.Text); err == nil && val > 0 {
		c.Config.Repetitions = val
	}
	if val, err := strconv.Atoi(c.DelayEntry.Text); err == nil && val >= 0 {
		c.Config.DelayMs = val
	}
	c.FlashController.UpdateConfig(c.Config)
}

// UpdatePorts updates the list of MIDI ports
func (c *Components) UpdatePorts() {
	ins, err := c.MIDIManager.GetInputPorts()
	if err != nil {
		return
	}
	outs, err := c.MIDIManager.GetOutputPorts()
	if err != nil {
		return
	}

	names := []string{}
	c.InPortsList.SetText("")
	for _, in := range ins {
		names = append(names, in.String())
		c.InPortsList.SetText(c.InPortsList.Text + in.String() + "\n")
	}
	c.MIDISelect.Options = names
	if len(names) > 0 {
		c.MIDISelect.SetSelected(names[0])
		c.startMIDI(ins[0])
	}

	c.OutPortsList.SetText("")
	for _, out := range outs {
		c.OutPortsList.SetText(c.OutPortsList.Text + out.String() + "\n")
	}
}

// createTabs creates the tab container
func (c *Components) createTabs() *container.AppTabs {
	flashTab := container.NewBorder(nil, nil, nil, nil, c.FlashContainer)
	configTab := c.createConfigTab()

	tabs := container.NewAppTabs(
		container.NewTabItem("Flash", flashTab),
		container.NewTabItem("Configuración", configTab),
	)
	tabs.SetTabLocation(container.TabLocationTop)
	return tabs
}

// createConfigTab creates the configuration tab content
func (c *Components) createConfigTab() *container.Scroll {
	updateFlashButton := widget.NewButton("Actualizar configuración de flash", func() {
		c.UpdateFlashConfig()
	})

	updatePortsButton := widget.NewButton("Actualizar lista de puertos", func() {
		c.UpdatePorts()
	})

	configContent := container.NewVBox(
		widget.NewLabel("Configuración del flash y MIDI"),
		widget.NewSeparator(),
		widget.NewLabel("Tiempo de flash (ms):"),
		c.FlashEntry,
		widget.NewLabel("Número de repeticiones:"),
		c.RepetitionsEntry,
		widget.NewLabel("Delay entre repeticiones (ms):"),
		c.DelayEntry,
		updateFlashButton,
		widget.NewLabel("Color del flash:"),
		c.ColorSelect,
		widget.NewSeparator(),
		widget.NewLabel("Selecciona puerto MIDI de entrada:"),
		c.MIDISelect,
		updatePortsButton,
		widget.NewSeparator(),
		widget.NewLabel("Puertos MIDI de entrada:"),
		c.InPortsList,
		widget.NewLabel("Puertos MIDI de salida:"),
		c.OutPortsList,
	)
	configScroll := container.NewVScroll(configContent)
	configScroll.SetMinSize(fyne.NewSize(300, 100))
	return configScroll
}
