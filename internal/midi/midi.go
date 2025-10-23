package midi

import (
	"fmt"
	"log"
	"time"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

// Manager handles MIDI driver and port management
type Manager struct {
	driver      *rtmididrv.Driver
	ins         []drivers.In
	currentPort string
	stopFunc    func()
}

// NewManager creates a new MIDI manager
func NewManager() (*Manager, error) {
	drv, err := rtmididrv.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create MIDI driver: %w", err)
	}

	return &Manager{
		driver: drv,
	}, nil
}

// GetInputPorts returns all available MIDI input ports
func (m *Manager) GetInputPorts() ([]drivers.In, error) {
	ins, err := m.driver.Ins()
	if err != nil {
		return nil, fmt.Errorf("failed to get input ports: %w", err)
	}
	m.ins = ins
	return ins, nil
}

// GetOutputPorts returns all available MIDI output ports
func (m *Manager) GetOutputPorts() ([]drivers.Out, error) {
	outs, err := m.driver.Outs()
	if err != nil {
		return nil, fmt.Errorf("failed to get output ports: %w", err)
	}
	return outs, nil
}

// MessageHandler is called when a MIDI message is received
type MessageHandler func(message string)

// ListenToPort starts listening to a MIDI input port
func (m *Manager) ListenToPort(in drivers.In, onMessage MessageHandler, onFlash func()) error {
	if in.String() == m.currentPort {
		return nil // Already listening to this port
	}
	m.currentPort = in.String()

	// Stop previous listener if any
	if m.stopFunc != nil {
		m.stopFunc()
		m.stopFunc = nil
		time.Sleep(200 * time.Millisecond)
	}

	// Open the port
	if err := in.Open(); err != nil {
		return fmt.Errorf("failed to open MIDI port: %w", err)
	}

	// Create channels for flash and messages
	flashCh := make(chan struct{}, 1)
	msgCh := make(chan string, 50)

	// Start listening
	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var ch, key, vel uint8
		if msg.GetNoteStart(&ch, &key, &vel) { // only NoteOn
			message := fmt.Sprintf("NoteOn: %s Ch:%d Vel:%d", midi.Note(key), ch, vel)

			// Trigger flash (non-blocking)
			select {
			case flashCh <- struct{}{}:
			default:
			}

			// Send message (non-blocking)
			select {
			case msgCh <- message:
			default:
			}

			fmt.Println(message)
		}
	}, midi.UseSysEx())

	if err != nil {
		return fmt.Errorf("failed to listen to MIDI: %w", err)
	}

	m.stopFunc = stop

	// Goroutine to handle flash events
	go func() {
		for range flashCh {
			if onFlash != nil {
				onFlash()
			}
		}
	}()

	// Goroutine to handle message events
	go func() {
		for msg := range msgCh {
			if onMessage != nil {
				onMessage(msg)
			}
		}
	}()

	return nil
}

// Stop stops the current MIDI listener
func (m *Manager) Stop() {
	if m.stopFunc != nil {
		m.stopFunc()
		m.stopFunc = nil
	}
}

// Close closes the MIDI driver
func (m *Manager) Close() {
	m.Stop()
	if m.driver != nil {
		log.Println("Closing MIDI driver")
	}
}
