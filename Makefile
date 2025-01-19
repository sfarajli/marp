PREFIX ?= /usr/local

marp: src/main.go
	go build -o marp ./src

install: marp
	mkdir -p ${DESTDIR}${PREFIX}/bin
	cp -f marp ${DESTDIR}${PREFIX}/bin
	cp -f lib/std.marp ${DESTDIR}${PREFIX}/bin

uninstall:
	rm -f ${DESTDIR}${PREFIX}/bin/marp

clean:
	rm -f marp

.PHONY: test clean
