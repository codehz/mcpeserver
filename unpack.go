package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/cheggaaa/pb.v1"
)

func extractFile(base string) {
	f, err := os.Open(base + ".tar.gz")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	gzf, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}
	tarReader := tar.NewReader(gzf)
	err = os.MkdirAll(base, os.ModePerm)
	if err != nil {
		panic(err)
	}
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		filename := header.Name
		target, err := os.OpenFile(filepath.Join(base, filename), os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
		if err != nil {
			panic(err)
		}
		bar := pb.StartNew(int(header.Size))
		bar.SetUnits(pb.U_BYTES_DEC)
		bar.SetRefreshRate(time.Millisecond * 20)
		bar.Start()
		bar.Prefix(fmt.Sprintf("%-20s", filename))
		_, err = io.Copy(target, bar.NewProxyReader(tarReader))
		if err != nil {
			panic(err)
		}
		bar.Finish()
	}
}

func handleDownload(registry string, base string, force bool) {
	printInfo("Authorizing...")
	token := auth()
	printPair("Token", token)
	printInfo("Fetching...")
	blob := getLayer(registry, token)
	printPair("Blob", blob)
	if checkVersion(base+".version", []byte(blob)) || force {
		printInfo("Download Binary...\033[0;31m")
		download(registry, token, blob, base+".tar.gz")
		fmt.Print("\033[0;31m")
		extractFile(base)
	} else {
		printWarn("Version Matched, Skipped Download.")
	}
}

func unpack(base string, file string) {
	r, err := zip.OpenReader(file)
	if err != nil {
		panic(err)
	}
	defer r.Close()
	os.MkdirAll(base, os.ModePerm)
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
		err = os.MkdirAll(filepath.Dir(filepath.Join(base, targetName)), os.ModePerm)
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
