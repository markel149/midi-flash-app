GOOS=windows GOARCH=amd64 go build -o midi-flash.exe main.go
GOOS=linux   GOARCH=amd64 go build -o midi-flash-linux main.go
GOOS=darwin  GOARCH=amd64 go build -o midi-flash-mac main.go
GOOS=darwin GOARCH=arm64 go build -o midi-flash-mac-arm main.go
