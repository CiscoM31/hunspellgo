.DEFAULT_GOAL := hunspellgo

.PHONY: hunspell
hunspell:
	[ -d hunspell ] || git -c advice.detachedHead=false clone https://github.com/hunspell/hunspell.git --branch v1.7.1 --single-branch && \
	set -ex && \
	cd hunspell && \
	autoreconf -vfi && \
	./configure --prefix=/usr/local && \
	make -j $(nproc)

.PHONY: hunspellgo
hunspellgo: hunspell install
	go build ./...
	go test ./...

.PHONY: install
install: hunspell
	cd hunspell && \
	sudo make install

.PHONY: clean
clean:
	rm -rf hunspell
