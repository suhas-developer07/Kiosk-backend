.PHONY: build build-KioskBackendFunction clean

build: build-KioskBackendFunction

build-KioskBackendFunction:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap main.go

clean:
	rm -f bootstrap
