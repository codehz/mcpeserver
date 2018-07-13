package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/cheggaaa/pb.v1"
)

func unpack(base string, file string) {
	r, err := zip.OpenReader(file)
	if err != nil {
		panic(err)
	}
	defer r.Close()
	os.MkdirAll(base, 0755)
	count := 0
	skip := 0
	for _, f := range r.File {
		targetName := f.Name
		if targetName == "lib/x86/libfmod.so" ||
			strings.HasPrefix(targetName, "res/") ||
			strings.HasPrefix(targetName, "org/") ||
			strings.HasPrefix(targetName, "assets/shaders/") ||
			strings.HasPrefix(targetName, "assets/skin_packs/") ||
			strings.HasPrefix(targetName, "assets/renderer/") ||
			strings.HasPrefix(targetName, "assets/animation/") ||
			strings.HasPrefix(targetName, "META-INF/") ||
			strings.HasSuffix(targetName, ".png") ||
			strings.HasSuffix(targetName, ".fsb") ||
			strings.HasSuffix(targetName, ".ttf") ||
			strings.HasSuffix(targetName, ".jpg") ||
			strings.HasSuffix(targetName, ".txt") ||
			strings.HasSuffix(targetName, ".tga") ||
			!strings.ContainsRune(targetName, '/') {
			skip++
			continue
		}
		if targetName == "lib/x86/libminecraftpe.so" {
			targetName = "libs/libminecraftpe.so"
		}
		count++
		fmt.Printf("\033[0;92m[%4d|%4d skiped]extracting: %s\033[0;31m", count, skip, targetName)
		err = os.MkdirAll(filepath.Dir(filepath.Join(base, targetName)), 0755)
		if err != nil {
			panic(err)
		}
		target, err := os.OpenFile(filepath.Join(base, targetName), os.O_CREATE|os.O_RDWR, os.FileMode(f.Mode()))
		if err != nil {
			panic(err)
		}
		rc, err := f.Open()
		if err != nil {
			panic(err)
		}
		defer rc.Close()
		time.Sleep(time.Millisecond) // Avoid some glitch about terminal
		if f.UncompressedSize >= 1024000 {
			fmt.Println()
			bar := pb.StartNew(int(f.UncompressedSize))
			bar.SetUnits(pb.U_BYTES_DEC)
			bar.SetRefreshRate(time.Millisecond * 20)
			bar.Start()
			_, err := io.Copy(target, bar.NewProxyReader(rc))
			if err != nil {
				panic(err)
			}
			bar.Finish()
		} else {
			fmt.Print("\033[1K\r")
			_, err := io.Copy(target, rc)
			if err != nil {
				panic(err)
			}
		}
	}
	fmt.Print("\033[0m")
}
