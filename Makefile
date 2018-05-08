mcpeserver: main.go
	GOARCH=386 go build -ldflags="-s -w"
	goupx --brute mcpeserver