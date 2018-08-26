package main

import (
	"os"
	"os/exec"
)

func runDaemon(datapath, profile string) {
	cmd := exec.Command("./bin/bedrockserver")
	cmd.Dir, _ = os.Getwd()
	cmd.Start()
}
