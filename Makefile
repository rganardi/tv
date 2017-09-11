VERSION:=$(shell git describe --tags --long --always)
BUILDDATE:=$(shell date "+%FT%T%z")
LDFLAGS=-ldflags "-X main.version_number=${VERSION} -X main.build_date=${BUILDDATE}"

tv: tv.go
	go build ${LDFLAGS}

install:
	go install ${LDFLAGS}

complete:
	install -m 644 _tv /usr/share/zsh/site-functions
