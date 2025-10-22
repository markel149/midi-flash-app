package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	defer midi.CloseDriver()

	a := app.New()
	w := a.NewWindow("MIDI Flash")
	w.Resize(fyne.NewSize(100, 100))
	w.SetPadded(false)

	// --- RECTÁNGULO DE FLASH ---
	flashRect := canvas.NewRectangle(color.Black)
	flashRect.SetMinSize(fyne.NewSize(100, 100))
	flashContainer := container.NewMax(flashRect)

	// --- MENSAJES MIDI ---
	msgBox := container.NewVBox()
	scrollMessages := container.NewVScroll(msgBox)
	scrollMessages.SetMinSize(fyne.NewSize(350, 100))

	// --- CONFIGURACIÓN ---
	inPortsList := widget.NewMultiLineEntry()
	inPortsList.SetPlaceHolder("Puertos MIDI detectados...")
	inPortsList.Disable()

	outPortsList := widget.NewMultiLineEntry()
	outPortsList.SetPlaceHolder("Puertos MIDI de salida...")
	outPortsList.Disable()

	flashTimeMs := 100
	repetitions := 1
	delayMs := 50

	flashEntry := widget.NewEntry()
	flashEntry.SetText(strconv.Itoa(flashTimeMs))

	repetitionsEntry := widget.NewEntry()
	repetitionsEntry.SetText(strconv.Itoa(repetitions))

	delayEntry := widget.NewEntry()
	delayEntry.SetText(strconv.Itoa(delayMs))

	// --- Selector de color ---
	flashColor := color.RGBA{R: 255, G: 0, B: 0, A: 255} // rojo por defecto
	colorSelect := widget.NewSelect([]string{"Rojo", "Verde", "Azul", "Blanco", "Amarillo", "Cian", "Magenta"}, func(s string) {
		switch s {
		case "Rojo":
			flashColor = color.RGBA{R: 255, G: 0, B: 0, A: 255}
		case "Verde":
			flashColor = color.RGBA{R: 0, G: 255, B: 0, A: 255}
		case "Azul":
			flashColor = color.RGBA{R: 0, G: 0, B: 255, A: 255}
		case "Blanco":
			flashColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
		case "Amarillo":
			flashColor = color.RGBA{R: 255, G: 255, B: 0, A: 255}
		case "Cian":
			flashColor = color.RGBA{R: 0, G: 255, B: 255, A: 255}
		case "Magenta":
			flashColor = color.RGBA{R: 255, G: 0, B: 255, A: 255}
		}
	})
	colorSelect.SetSelected("Rojo")

	var stop func()
	var drv *rtmididrv.Driver
	var err error
	var ins []drivers.In
	var currentPort string

	startMIDI := func(in drivers.In) {
		if in.String() == currentPort {
			return
		}
		currentPort = in.String()

		if stop != nil {
			stop()
			stop = nil
			time.Sleep(200 * time.Millisecond)
		}

		if err := in.Open(); err != nil {
			log.Printf("No se pudo abrir puerto MIDI: %v", err)
			return
		}

		flashCh := make(chan struct{}, 1)
		msgCh := make(chan string, 50)

		stop, err = midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
			var ch, key, vel uint8
			if msg.GetNoteStart(&ch, &key, &vel) { // solo NoteOn
				message := fmt.Sprintf("NoteOn: %s Ch:%d Vel:%d", midi.Note(key), ch, vel)

				select {
				case flashCh <- struct{}{}:
				default:
				}

				select {
				case msgCh <- message:
				default:
				}

				fmt.Println(message)
			}
		}, midi.UseSysEx())

		if err != nil {
			log.Printf("Error al escuchar MIDI: %v", err)
			return
		}

		// Goroutine para flash
		go func() {
			for range flashCh {
				for i := 0; i < repetitions; i++ {
					fyne.Do(func() {
						flashRect.FillColor = flashColor
						flashRect.Refresh()
					})
					time.Sleep(time.Duration(flashTimeMs) * time.Millisecond)
					fyne.Do(func() {
						flashRect.FillColor = color.Black
						flashRect.Refresh()
					})
					time.Sleep(time.Duration(delayMs) * time.Millisecond)
				}
			}
		}()

		// Mostrar mensajes
		go func() {
			for msg := range msgCh {
				m := msg
				fyne.Do(func() {
					msgLabel := widget.NewLabel(m)
					msgBox.Add(msgLabel)
					scrollMessages.ScrollToBottom()
				})
			}
		}()
	}

	updateFlashButton := widget.NewButton("Actualizar configuración de flash", func() {
		if val, err := strconv.Atoi(flashEntry.Text); err == nil && val > 0 {
			flashTimeMs = val
		}
		if val, err := strconv.Atoi(repetitionsEntry.Text); err == nil && val > 0 {
			repetitions = val
		}
		if val, err := strconv.Atoi(delayEntry.Text); err == nil && val >= 0 {
			delayMs = val
		}
	})

	// Inicializa driver
	drv, err = rtmididrv.New()
	if err != nil {
		log.Fatal(err)
	}

	midiSelect := widget.NewSelect([]string{}, func(selected string) {
		for _, in := range ins {
			if in.String() == selected {
				startMIDI(in)
				break
			}
		}
	})

	updatePorts := func() {
		ins, err = drv.Ins()
		if err != nil {
			log.Fatal(err)
		}
		outs, err := drv.Outs()
		if err != nil {
			log.Fatal(err)
		}

		names := []string{}
		inPortsList.SetText("")
		for _, in := range ins {
			names = append(names, in.String())
			inPortsList.SetText(inPortsList.Text + in.String() + "\n")
		}
		midiSelect.Options = names
		if len(names) > 0 {
			midiSelect.SetSelected(names[0])
			startMIDI(ins[0])
		}

		outPortsList.SetText("")
		for _, out := range outs {
			outPortsList.SetText(outPortsList.Text + out.String() + "\n")
		}
	}

	updatePortsButton := widget.NewButton("Actualizar lista de puertos", func() {
		updatePorts()
	})

	// --- Pestaña de configuración con scroll ---
	configContent := container.NewVBox(
		widget.NewLabel("Configuración del flash y MIDI"),
		widget.NewSeparator(),
		widget.NewLabel("Tiempo de flash (ms):"),
		flashEntry,
		widget.NewLabel("Número de repeticiones:"),
		repetitionsEntry,
		widget.NewLabel("Delay entre repeticiones (ms):"),
		delayEntry,
		updateFlashButton,
		widget.NewLabel("Color del flash:"),
		colorSelect,
		widget.NewSeparator(),
		widget.NewLabel("Selecciona puerto MIDI de entrada:"),
		midiSelect,
		updatePortsButton,
		widget.NewSeparator(),
		widget.NewLabel("Puertos MIDI de entrada:"),
		inPortsList,
		widget.NewLabel("Puertos MIDI de salida:"),
		outPortsList,
	)
	configScroll := container.NewVScroll(configContent)
	configScroll.SetMinSize(fyne.NewSize(300, 100)) // tamaño inicial mínimo

	flashTab := container.NewBorder(nil, nil, nil, nil, flashContainer)

	tabs := container.NewAppTabs(
		container.NewTabItem("Flash", flashTab),
		container.NewTabItem("Configuración", configScroll),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		if stop != nil {
			stop()
		}
		a.Quit()
	}()

	updatePorts()

	w.SetContent(tabs)
	w.ShowAndRun()
}
