package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type modsRepoInfo struct {
	Version string   `json:"verion"`
	List    []string `json:"list"`
}

type modInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Main    string `json:"main"`
	Info    map[string]struct {
		Name string `json:"name"`
		Desc string `json:"description"`
	}
}

func listLocalMod() {
	files, err := ioutil.ReadDir("./games/mods")
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".so") {
			printInfo(f.Name())
		}
	}
}

func listRemoteMod(endpoint string) {
	resp, err := http.Get(endpoint + "/mods")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	info := modsRepoInfo{}
	if err = json.Unmarshal(contents, &info); err != nil {
		panic(err)
	}
	for _, item := range info.List {
		printInfo(item)
	}
}

func infoRemoteMod(endpoint, name string) {
	resp, err := http.Get(endpoint + "/mods/" + name)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	info := modInfo{}
	if err = json.Unmarshal(contents, &info); err != nil {
		panic(err)
	}
	lang := os.Getenv("LANG")
	if len(lang) < 5 {
		lang = "en-US"
	} else {
		lang = strings.Replace(lang[0:5], "_", "-", 1)
	}
	linfo, ok := info.Info[lang]
	if !ok {
		linfo = info.Info["en-US"]
	}
	printPair("Name", linfo.Name)
	printPair("Desc", linfo.Desc)
	printPair("Version", info.Version)
	printPair("Main", info.Main)
}

func downloadMod(endpoint, name string) {
	printInfo("Fetch metadata...")
	resp, err := http.Get(endpoint + "/mods/" + name)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	info := modInfo{}
	if err = json.Unmarshal(contents, &info); err != nil {
		panic(err)
	}
	printPair("Main", info.Main)
	printInfo("Downloading...")
	target := "games/mods/" + info.Main
	out, err := os.OpenFile(target+".tmp", os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}
	os.Rename(target+".tmp", target)
	printInfo("Finished.")
}
