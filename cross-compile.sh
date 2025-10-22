GOOS=windows GOARCH=amd64 go build -o midi-flash.exe main.go
GOOS=linux   GOARCH=amd64 go build -o midi-flash-linux main.go
# Para Intel (amd64)
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o midi-flash-mac main.go

# Para Apple Silicon (arm64)
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o midi-flash-mac-arm main.go