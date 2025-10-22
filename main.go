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
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	defer midi.CloseDriver()

	a := app.New()
	w := a.NewWindow("MIDI Flash")
	w.Resize(fyne.NewSize(1000, 600))
	w.SetPadded(false)

	// --- FLASH ---
	flashRect := canvas.NewRectangle(color.White)
	flashRect.SetMinSize(fyne.NewSize(600, 400))
	flashContainer := container.NewMax(flashRect)

	// --- MENSAJES MIDI ---
	msgBox := container.NewVBox()
	scrollMessages := container.NewVScroll(msgBox)
	scrollMessages.SetMinSize(fyne.NewSize(350, 400))

	// --- CONFIGURACIÓN ---
	inPortsList := widget.NewMultiLineEntry()
	inPortsList.SetPlaceHolder("Puertos MIDI de entrada...")
	inPortsList.Disable()

	outPortsList := widget.NewMultiLineEntry()
	outPortsList.SetPlaceHolder("Puertos MIDI de salida...")
	outPortsList.Disable()

	flashTimeMs := 100
	flashEntry := widget.NewEntry()
	flashEntry.SetText(strconv.Itoa(flashTimeMs))
	flashEntry.SetPlaceHolder("Tiempo de flash en ms")

	flashButton := widget.NewButton("Actualizar tiempo de flash", func() {
		if val, err := strconv.Atoi(flashEntry.Text); err == nil && val > 0 {
			flashTimeMs = val
		}
	})

	configTab := container.NewVBox(
		widget.NewLabel("Configuración del flash y MIDI"),
		widget.NewSeparator(),
		widget.NewLabel("Tiempo de flash (ms):"),
		flashEntry,
		flashButton,
		widget.NewSeparator(),
		widget.NewLabel("Puertos MIDI de entrada:"),
		inPortsList,
		widget.NewLabel("Puertos MIDI de salida:"),
		outPortsList,
	)

	// --- INICIALIZA DRIVER MIDI ---
	drv, err := rtmididrv.New()
	if err != nil {
		log.Fatal(err)
	}
	defer drv.Close()

	ins, err := drv.Ins()
	if err != nil {
		log.Fatal(err)
	}
	inPortsList.SetText("")
	for i, in := range ins {
		inPortsList.SetText(inPortsList.Text + fmt.Sprintf("[%d] %s\n", i, in.String()))
	}

	outs, err := drv.Outs()
	if err != nil {
		log.Fatal(err)
	}
	outPortsList.SetText("")
	for i, out := range outs {
		outPortsList.SetText(outPortsList.Text + fmt.Sprintf("[%d] %s\n", i, out.String()))
	}

	// --- SELECCIONA PRIMER PUERTO MIDI ---
	in, err := midi.InPort(0)
	if err != nil {
		fmt.Println("No se encuentra puerto de entrada MIDI (usa índice válido)")
		return
	}

	flashCh := make(chan struct{}, 1)
	msgCh := make(chan string, 50)

	// --- LISTENER MIDI ---
	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		select {
		case flashCh <- struct{}{}:
		default:
		}

		var bt []byte
		var ch, key, vel uint8
		message := ""
		switch {
		case msg.GetSysEx(&bt):
			message = fmt.Sprintf("SysEx: % X", bt)
		case msg.GetNoteStart(&ch, &key, &vel):
			message = fmt.Sprintf("NoteOn: %s Ch:%d Vel:%d", midi.Note(key), ch, vel)
		case msg.GetNoteEnd(&ch, &key):
			message = fmt.Sprintf("NoteOff: %s Ch:%d", midi.Note(key), ch)
		default:
			message = fmt.Sprintf("Other: % X", msg.Bytes())
		}

		select {
		case msgCh <- message:
		default:
		}

		fmt.Println(message)
	}, midi.UseSysEx())

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	// --- GOROUTINE FLASH ---
	go func() {
		for range flashCh {
			fyne.Do(func() {
				flashRect.FillColor = color.RGBA{R: 255, G: 0, B: 0, A: 255}
				flashRect.Refresh()
			})
			time.Sleep(time.Duration(flashTimeMs) * time.Millisecond)
			fyne.Do(func() {
				flashRect.FillColor = color.White
				flashRect.Refresh()
			})
		}
	}()

	// --- GOROUTINE MENSAJES ---
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

	// --- MANEJO CTRL+C ---
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		stop()
		a.Quit()
	}()

	// --- TABS ---
	flashTab := container.NewBorder(nil, nil, nil, nil, flashContainer)
	tabs := container.NewAppTabs(
		container.NewTabItem("Flash", flashTab),
		container.NewTabItem("Configuración", configTab),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	w.SetContent(tabs)
	w.ShowAndRun()
}
