.PHONY: all deps build package install clean run

ARCH ?= amd64

all: deps build package

deps:
	./deps.sh

build:
	mkdir -p ./deb/opt/wireguird/
	GOMAXPROCS=$$(nproc) go generate
	GOMAXPROCS=$$(nproc) GOARCH=$(ARCH) go build -ldflags "-s -w" -trimpath -o ./deb/opt/wireguird/wireguird

package:
	cp -r ./Icon/ ./deb/opt/wireguird/
	echo '{"MultipleTunnels":false,"StartOnTray":false,"CheckUpdates":false,"TunnelsPath":"/etc/wireguard","Debug":false}' > ./deb/opt/wireguird/wireguird.settings
	mkdir -p ./build/
	sed -i "s/Architecture: .*/Architecture: $(ARCH)/" ./deb/DEBIAN/control
	dpkg-deb --root-owner-group --build ./deb ./build/wireguird_$(ARCH).deb

install:
	sudo dpkg -i ./build/wireguird_$(ARCH).deb

clean:
	rm -rf ./build/
	rm -rf ./deb/opt/wireguird/

run:
	go generate
	go build
	sudo ./wireguird