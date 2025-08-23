set windows-powershell := true

ui-run:
	go mod tidy
	go run -tags production,desktop -ldflags "-w -h -H windowsgui" .\...\0xwaraqah

ui-build:
	go mod tidy
	go build -tags production,desktop -ldflags "-w -h -H windowsgui" .\...\0xwaraqah
