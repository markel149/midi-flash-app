package main

import (
	"log"
	"os"
	"os/signal"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"gitlab.com/gomidi/midi/v2"

	midimanager "github.com/markel149/midi-flash-app/internal/midi"
	"github.com/markel149/midi-flash-app/internal/ui"
)

func main() {
	defer midi.CloseDriver()

	// Initialize MIDI manager
	midiManager, err := midimanager.NewManager()
	if err != nil {
		log.Fatal(err)
	}
	defer midiManager.Close()

	// Create Fyne application
	a := app.New()
	w := a.NewWindow("MIDI Flash")
	w.Resize(fyne.NewSize(100, 100))
	w.SetPadded(false)

	// Create UI components
	components := ui.NewComponents(midiManager)

	// Handle interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		midiManager.Stop()
		a.Quit()
	}()

	// Initialize MIDI ports
	components.UpdatePorts()

	// Set window content and run
	w.SetContent(components.Tabs)
	w.ShowAndRun()
}
