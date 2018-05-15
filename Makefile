.PHONY: release all

all: release

release: mcpeserver
	goupx --brute mcpeserver

mcpeserver: ${wildcard *.go}
	GOARCH=386 go build -ldflags="-s -w"