# install dependencies and compile server

VERSION ?= latest

default:
	make it

clean:
	rm ./server

.PHONY: build test dist it

build:
	go get
	mkdir -p ./bin
	go build -o ./bin/lrs ./cmd/liverecord
	chmod a+x ./bin/lrs

it:
	make build
	if [ -f .env ]; then echo "Starting server!"; ./bin/lrs ; else echo "Create .env file! Check manual on github!" ; fi ;

test:
	./test.sh

dist:
	rm -rf dist
	mkdir -p build/alpine-linux/amd64 && GOOS=linux GOARCH=amd64 go build -a -tags netgo -installsuffix netgo -o build/alpine-linux/amd64/lrs .
	mkdir -p build/linux/amd64 && GOOS=linux GOARCH=amd64 go build -o build/linux/amd64/lrs .
	mkdir -p build/linux/i386 && GOOS=linux GOARCH=386 go build -o build/linux/i386/lrs .
	mkdir -p build/linux/armel && GOOS=linux GOARCH=arm GOARM=5 go build -o build/linux/armel/lrs .
	mkdir -p build/linux/armhf && GOOS=linux GOARCH=arm GOARM=6 go build -o build/linux/armhf/lrs .
	mkdir -p build/darwin/amd64 && GOOS=darwin GOARCH=amd64 go build -o build/darwin/amd64/lrs .
	mkdir -p build/darwin/i386 && GOOS=darwin GOARCH=386 go build -o build/darwin/i386/lrs .
	mkdir -p build/windows/i386 && GOOS=windows GOARCH=386 go build -o build/windows/i386/lrs.exe .
	mkdir -p build/windows/amd64 && GOOS=windows GOARCH=amd64 go build -o build/windows/amd64/lrs.exe .

	mkdir -p dist/

	tar -cvzf dist/lrs-alpine-linux-amd64-$(VERSION).tar.gz -C build/alpine-linux/amd64 lrs
	tar -cvzf dist/lrs-linux-amd64-$(VERSION).tar.gz -C build/linux/amd64 lrs
	tar -cvzf dist/lrs-linux-i386-$(VERSION).tar.gz -C build/linux/i386 lrs
	tar -cvzf dist/lrs-linux-armel-$(VERSION).tar.gz -C build/linux/armel lrs
	tar -cvzf dist/lrs-linux-armhf-$(VERSION).tar.gz -C build/linux/armhf lrs
	tar -cvzf dist/lrs-darwin-amd64-$(VERSION).tar.gz -C build/darwin/amd64 lrs
	zip dist/lrs-windows-i386-$(VERSION).zip build/windows/i386/lrs.exe
	zip dist/lrs-windows-amd64-$(VERSION).zip build/windows/amd64/lrs.exe
	rm -rf build
