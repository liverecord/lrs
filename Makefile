# install dependencies and compile server

default:
	make build

clean:
	rm ./server

VERSION ?= latest

.PHONY: build test dist

build:
	mkdir -p ./bin
	go build -o ./bin/liverecord-server .

test:
	./test.sh

dist:
	rm -rf dist
	mkdir -p build/alpine-linux/amd64 && GOOS=linux GOARCH=amd64 go build -a -tags netgo -installsuffix netgo -o build/alpine-linux/amd64/liverecord-server .
	mkdir -p build/linux/amd64 && GOOS=linux GOARCH=amd64 go build -o build/linux/amd64/liverecord-server .
	mkdir -p build/linux/i386 && GOOS=linux GOARCH=386 go build -o build/linux/i386/liverecord-server .
	mkdir -p build/linux/armel && GOOS=linux GOARCH=arm GOARM=5 go build -o build/linux/armel/liverecord-server .
	mkdir -p build/linux/armhf && GOOS=linux GOARCH=arm GOARM=6 go build -o build/linux/armhf/liverecord-server .
	mkdir -p build/darwin/amd64 && GOOS=darwin GOARCH=amd64 go build -o build/darwin/amd64/liverecord-server .
	mkdir -p build/darwin/i386 && GOOS=darwin GOARCH=386 go build -o build/darwin/i386/liverecord-server .
	mkdir -p build/windows/i386 && GOOS=windows GOARCH=386 go build -o build/windows/i386/liverecord-server.exe .
	mkdir -p build/windows/amd64 && GOOS=windows GOARCH=amd64 go build -o build/windows/amd64/liverecord-server.exe .

	mkdir -p dist/

	tar -cvzf dist/liverecord-server-alpine-linux-amd64-$(VERSION).tar.gz -C build/alpine-linux/amd64 liverecord-server
	tar -cvzf dist/liverecord-server-linux-amd64-$(VERSION).tar.gz -C build/linux/amd64 liverecord-server
	tar -cvzf dist/liverecord-server-linux-i386-$(VERSION).tar.gz -C build/linux/i386 liverecord-server
	tar -cvzf dist/liverecord-server-linux-armel-$(VERSION).tar.gz -C build/linux/armel liverecord-server
	tar -cvzf dist/liverecord-server-linux-armhf-$(VERSION).tar.gz -C build/linux/armhf liverecord-server
	tar -cvzf dist/liverecord-server-darwin-amd64-$(VERSION).tar.gz -C build/darwin/amd64 liverecord-server
	zip dist/liverecord-server-windows-i386-$(VERSION).zip build/windows/i386/liverecord-server.exe
	zip dist/liverecord-server-windows-amd64-$(VERSION).zip build/windows/amd64/liverecord-server.exe
	rm -rf build
