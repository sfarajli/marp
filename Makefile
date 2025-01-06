PREFIX ?= /usr/local

ash: src/main.go
	go build -o ash ./src

install: ash
	mkdir -p ${DESTDIR}${PREFIX}/bin
	cp -f ash ${DESTDIR}${PREFIX}/bin

uninstall:
	rm -f ${DESTDIR}${PREFIX}/bin/ash

clean:
	rm -f ash

.PHONY: test clean
