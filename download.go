package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"gopkg.in/cheggaaa/pb.v1"
)

type authResp struct {
	Token string `json:"token"`
}

type manifest struct {
	FsLayers []struct {
		BlobSum string `json:"blobSum"`
	} `json:"fsLayers"`
}

func auth() string {
	resp, err := http.Get("https://auth.docker.io/token?service=registry.docker.io&scope=repository:codehz/bedrockserver:pull")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	auth := authResp{}
	if err := json.Unmarshal(contents, &auth); err != nil {
		panic(err)
	}
	return auth.Token
}

func fetch(url string, token string) []byte {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return contents
}

func getLayer(registry string, token string) string {
	contents := fetch(registry+"/codehz/bedrockserver/manifests/latest", token)
	info := manifest{}
	if err := json.Unmarshal(contents, &info); err != nil {
		panic(err)
	}
	return info.FsLayers[0].BlobSum
}

func download(registry string, token string, blob string, target string) {
	out, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	client := &http.Client{}
	req, err := http.NewRequest("GET", registry+"/codehz/bedrockserver/blobs/"+blob, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	bar := pb.StartNew(int(resp.ContentLength))
	bar.SetUnits(pb.U_BYTES_DEC)
	bar.SetRefreshRate(time.Millisecond * 20)
	bar.Prefix(fmt.Sprintf("%-20s", target))
	bar.Start()
	_, err = io.Copy(out, bar.NewProxyReader(resp.Body))
	if err != nil {
		panic(err)
	}
	bar.FinishPrint("\033[0;34mDownload Finished.\033[0m")
}

func checkVersion(filepath string, blob []byte) bool {
	if _, err := os.Stat(filepath); err == nil {
		content, err := ioutil.ReadFile(filepath)
		if err != nil {
			panic(err)
		}
		if bytes.Compare(content, blob) == 0 {
			return false
		}
	}
	err := ioutil.WriteFile(filepath, []byte(blob), 0644)
	if err != nil {
		panic(err)
	}
	return true
}

func extractFile() {
	f, err := os.Open("bin.tar.gz")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	gzf, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}
	tarReader := tar.NewReader(gzf)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		filename := header.Name
		target, err := os.OpenFile(path.Join("bin", filename), os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
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

func handleDownload(registry string, force bool) {
	printInfo("Authorizing...")
	token := auth()
	printPair("Token", token)
	printInfo("Fetching...")
	blob := getLayer(registry, token)
	printPair("Blob", blob)
	if checkVersion(".version", []byte(blob)) || force {
		printInfo("Download Binary...\033[0;31m")
		download(registry, token, blob, "bin.tar.gz")
		fmt.Print("\033[0;31m")
		os.MkdirAll("bin", 0755)
		extractFile()
		fmt.Print("\033[0m")
	} else {
		printWarn("Version Matched, Skipped Download.")
	}
}
