package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"gopkg.in/cheggaaa/pb.v1"
)

type releaseInfo struct {
	Assets []struct {
		URL string `json:"browser_download_url"`
	} `json:"assets"`
}

func getServerURL() string {
	resp, err := http.Get("https://api.github.com/repos/codehz/mcpeserver/releases/latest")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	info := releaseInfo{}
	if err = json.Unmarshal(contents, &info); err != nil {
		panic(err)
	}
	return info.Assets[0].URL
}

func fetchBinary(url string, target string) {
	out, err := os.OpenFile(target+".tmp", os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bar := pb.StartNew(int(resp.ContentLength))
	bar.SetUnits(pb.U_BYTES_DEC)
	bar.SetRefreshRate(time.Millisecond * 20)
	bar.Prefix("mcpeserver")
	bar.Start()
	_, err = io.Copy(out, bar.NewProxyReader(resp.Body))
	if err != nil {
		panic(err)
	}
	bar.FinishPrint("\033[0;34mUpdate Finished.\033[0m")
	os.Rename(target+".tmp", target)
}
