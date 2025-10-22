# midi-flash-app
'''
export CGO_CFLAGS="-mmacosx-version-min=14.0"                                
export CGO_LDFLAGS="-mmacosx-version-min=14.0"
GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -o midi-flash-mac-arm main.go
'''