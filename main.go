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
	w.Resize(fyne.NewSize(1000, 600)) // ventana grande y redimensionable
	w.SetPadded(false)

	// Rectángulo que hace flash
	flashRect := canvas.NewRectangle(color.White)
	flashRect.SetMinSize(fyne.NewSize(600, 400)) // espacio grande para flash
	flashContainer := container.NewMax(flashRect)

	// Panel de mensajes MIDI (derecha) con scroll funcional
	msgBox := container.NewVBox() // VBox donde se añaden los mensajes
	scrollMessages := container.NewVScroll(msgBox)
	scrollMessages.SetMinSize(fyne.NewSize(350, 400))

	// Listas de puertos
	inPortsList := widget.NewMultiLineEntry()
	inPortsList.SetPlaceHolder("Puertos MIDI de entrada...")
	inPortsList.Disable()
	outPortsList := widget.NewMultiLineEntry()
	outPortsList.SetPlaceHolder("Puertos MIDI de salida...")
	outPortsList.Disable()

	// Flash time configurable
	flashTimeMs := 100
	flashEntry := widget.NewEntry()
	flashEntry.SetText(strconv.Itoa(flashTimeMs))
	flashEntry.SetPlaceHolder("Tiempo de flash en ms")

	flashButton := widget.NewButton("Actualizar tiempo de flash", func() {
		if val, err := strconv.Atoi(flashEntry.Text); err == nil && val > 0 {
			flashTimeMs = val
		}
	})

	menu := container.NewVBox(
		widget.NewLabel("Tiempo de flash (ms):"),
		flashEntry,
		flashButton,
		widget.NewLabel("Puertos MIDI de entrada:"),
		inPortsList,
		widget.NewLabel("Puertos MIDI de salida:"),
		outPortsList,
	)

	// Panel principal: izquierda flash, derecha mensajes MIDI
	mainContent := container.NewHSplit(flashContainer, scrollMessages)
	mainContent.SetOffset(0.6) // 60% izquierda, 40% derecha

	// Contenedor final: menú arriba, contenido abajo
	mainContainer := container.NewBorder(menu, nil, nil, nil, mainContent)
	w.SetContent(mainContainer)

	// Inicializa driver MIDI
	drv, err := rtmididrv.New()
	if err != nil {
		log.Fatal(err)
	}
	defer drv.Close()

	// Listar puertos de entrada
	ins, err := drv.Ins()
	if err != nil {
		log.Fatal(err)
	}
	inPortsList.SetText("")
	for i, in := range ins {
		inPortsList.SetText(inPortsList.Text + fmt.Sprintf("[%d] %s\n", i, in.String()))
	}

	// Listar puertos de salida
	outs, err := drv.Outs()
	if err != nil {
		log.Fatal(err)
	}
	outPortsList.SetText("")
	fmt.Println(outs)
	for i, out := range outs {
		fmt.Println(i)
		fmt.Println(out)

		outPortsList.SetText(outPortsList.Text + fmt.Sprintf("[%d] %s\n", i, out.String()))
	}

	// Selecciona primer puerto
	in, err := midi.InPort(0)
	if err != nil {
		fmt.Println("can't find input port")
		return
	}

	flashCh := make(chan struct{}, 1)
	msgCh := make(chan string, 50) // canal para mensajes MIDI

	// Listener MIDI
	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		// señal de flash
		select {
		case flashCh <- struct{}{}:
		default:
		}

		// preparar mensaje de consola
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

		// imprimir en consola
		fmt.Println(message)
	}, midi.UseSysEx())

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	// Goroutine para manejar flash en hilo principal usando fyne.Do
	go func() {
		for range flashCh {
			fyne.Do(func() {
				flashRect.FillColor = color.RGBA{R: 255, G: 0, B: 0, A: 255} // rojo
				flashRect.Refresh()
			})

			time.Sleep(time.Duration(flashTimeMs) * time.Millisecond)

			fyne.Do(func() {
				flashRect.FillColor = color.White
				flashRect.Refresh()
			})
		}
	}()

	// Goroutine para actualizar panel de mensajes MIDI
	go func() {
		for msg := range msgCh {
			m := msg // captura variable
			fyne.Do(func() {
				msgLabel := widget.NewLabel(m)
				msgBox.Add(msgLabel) // añade nueva línea
				scrollMessages.ScrollToBottom()
			})
		}
	}()

	// Ctrl+C para salir
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		stop()
		a.Quit()
	}()

	w.ShowAndRun()
}
