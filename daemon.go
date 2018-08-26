package main

import (
	"os"
	"os/exec"
)

func runDaemon(profile string, systemd bool) {
	cmd := exec.Command("./bin/bedrockserver")
	cmd.Dir, _ = os.Getwd()
	cmd.Start()
	if systemd {
		cmd.Wait()
	}
}
