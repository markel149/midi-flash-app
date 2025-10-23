# midi-flash-app

A small desktop utility (GUI) that listens to a MIDI input port and flashes a rectangle on incoming NoteOn messages. It uses Fyne for the GUI and gomidi/rtmididrv for MIDI input.

## Project Structure

The application has been refactored into a modular structure for better maintainability and code reusability:

```
midi-flash-app/
├── main.go                    # Application entry point
├── internal/                  # Internal packages
│   ├── config/               # Configuration management
│   │   └── config.go         # Flash configuration and color handling
│   ├── flash/                # Flash controller
│   │   └── flash.go          # Flash rectangle behavior and timing
│   ├── midi/                 # MIDI handling
│   │   └── midi.go           # MIDI driver, port management, and message processing
│   └── ui/                   # User interface components
│       └── ui.go             # UI widgets, tabs, and event handlers
├── go.mod                    # Go module definition
├── go.sum                    # Go dependencies checksums
└── cross-compile.sh          # Cross-compilation helper script
```

### Package Overview

- **`main.go`**: Application entry point that initializes the MIDI manager, creates the UI, and handles application lifecycle.

- **`internal/config`**: Manages flash configuration including timing, repetitions, delay, and colors. Provides color lookup by name and default configuration values.

- **`internal/flash`**: Controls the flash rectangle behavior. Handles flash triggering, timing, repetitions, and color changes. Runs flash events asynchronously.

- **`internal/midi`**: Manages MIDI driver initialization, port enumeration, and message listening. Handles NoteOn events and provides callbacks for flash and message display.

- **`internal/ui`**: Creates and manages all UI components including the flash view, configuration panel, MIDI port selection, and message display. Coordinates between MIDI events and UI updates.

## Files
- [main.go](main.go) — application entrypoint. Initializes MIDI manager and UI components.
- [internal/config/config.go](internal/config/config.go) — flash configuration and color management.
- [internal/flash/flash.go](internal/flash/flash.go) — flash controller with timing and color control.
- [internal/midi/midi.go](internal/midi/midi.go) — MIDI driver management and message handling.
- [internal/ui/ui.go](internal/ui/ui.go) — UI components, tabs, and event coordination.
- [cross-compile.sh](cross-compile.sh) — convenience script with example cross-build commands.
- [go.mod](go.mod) — module and dependency list.
- [.gitignore](.gitignore)

## What the app does
- Creates a minimal Fyne GUI with two tabs:
  - "Flash": shows a rectangle that changes color briefly when a MIDI NoteOn is received.
  - "Configuración": lets you configure flash duration, repetitions, delay, and color; shows detected MIDI input/output ports and textual MIDI messages.
- Uses the RtMidi driver via `rtmididrv.New()` and listens to a selected input. When a NoteOn is detected it:
  - Triggers a non-blocking flash event that changes the rectangle color for the configured duration(s).
  - Appends a textual message in the UI showing note/channel/velocity.

## Build and run

1. Fetch dependencies
```sh
go mod tidy
```

2. Build for your current platform (recommended for development)
```sh
go build -o midi-flash main.go
./midi-flash
```

3. Cross-build examples
- The repository includes [cross-compile.sh](cross-compile.sh) with example commands:
  - Windows (amd64): GOOS=windows GOARCH=amd64 go build -o midi-flash.exe main.go
  - Linux (amd64):   GOOS=linux   GOARCH=amd64 go build -o midi-flash-linux main.go
  - macOS (amd64):   GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o midi-flash-mac main.go
  - macOS (arm64):   GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o midi-flash-mac-arm main.go

## Important notes about macOS / CGO and native MIDI drivers
- The RtMidi-based driver may require CGO and platform MIDI native libraries. Cross-compiling macOS binaries from a non-macOS host or without proper SDK headers can fail.
- If you build for macOS on macOS and need to set minimum macOS version, use:
```sh
export CGO_ENABLED=1
export CGO_CFLAGS="-mmacosx-version-min=14.0"
export CGO_LDFLAGS="-mmacosx-version-min=14.0"
GOOS=darwin GOARCH=arm64 go build -o midi-flash-mac-arm main.go
```
and similarly for `GOARCH=amd64`. These environment variables are only necessary when cgo is involved and you want to set an explicit macOS deployment target.

## Runtime requirements
- A working system MIDI backend (RtMidi/system MIDI) available to the driver. If the driver cannot enumerate or open ports the app will log errors.
- GUI: Fyne-supported environment (desktop with GPU/GL support).

If you run into build/link errors when cross-building for macOS, build on a macOS machine with CGO enabled, the correct SDK, or use a macOS CI runner.

## Development

The modular structure allows for easier testing and maintenance:

- **Adding new features**: Add new packages under `internal/` following the existing pattern.
- **Modifying MIDI handling**: Edit `internal/midi/midi.go`.
- **Changing flash behavior**: Edit `internal/flash/flash.go`.
- **Updating UI**: Edit `internal/ui/ui.go`.
- **Adding configuration options**: Edit `internal/config/config.go`.