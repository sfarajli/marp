PREFIX ?= /usr/local


gorth: src/main.go
	go build -o gorth ./src

install: gorth
	mkdir -p ${DESTDIR}${PREFIX}/bin
	cp -f gorth ${DESTDIR}${PREFIX}/bin

uninstall:
	rm -f ${DESTDIR}${PREFIX}/bin/gorth

clean:
	rm -f gorth

.PHONY: test clean
