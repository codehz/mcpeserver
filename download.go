package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
