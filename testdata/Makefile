default: all
.PHONY: default

.PHONY: clean
clean:
	rm -f windows-amd64.exe windows-386.exe linux-amd64 darwin-amd64 darwin-arm64

.PHONY: all
all: windows-amd64.exe windows-386.exe linux-amd64 darwin-amd64 darwin-arm64

.PHONY: windows-amd64.exe
windows-amd64.exe:
	GOOS=windows GOARCH=amd64 go build -o $@

.PHONY: windows-386.exe
windows-386.exe:
	GOOS=windows GOARCH=386 go build -o $@

.PHONY: linux-amd64
linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o $@

.PHONY: darwin-amd64
darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -o $@

.PHONY: darwin-arm64
darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o $@