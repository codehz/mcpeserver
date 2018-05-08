package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/subcommands"
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
	resp, err := http.Get("https://auth.docker.io/token?service=registry.docker.io&scope=repository:codehz/mcpe-server:pull")
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
	contents := fetch(registry+"/codehz/mcpe-server/manifests/latest", token)
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
	req, err := http.NewRequest("GET", registry+"/codehz/mcpe-server/blobs/"+blob, nil)
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

func printInfo(item string) {
	fmt.Printf("\033[0;32m%s\n", item)
}

func printWarn(item string) {
	fmt.Printf("\033[0;91m%s\n", item)
}

func printPair(key string, value string) {
	fmt.Printf("\033[0;34m%s: \033[0;35m%s\n", key, value)
}

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

func run(base string, datapath string) {
	abs, err := filepath.Abs(base)
	if err != nil {
		panic(err)
	}
	cmd := exec.Command(filepath.Join(abs, "server"))
	cmd.Env = append(os.Environ(), fmt.Sprintf("LD_LIBRARY_PATH=%s", abs))
	cmd.Dir = datapath
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		printWarn("Request Kill Server")
		cmd.Process.Kill()
	}()
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	cmd.Wait()
	signal.Reset(os.Interrupt)
	printInfo("Finished")
}

type downloadCmd struct {
	registry string
	target   string
	force    bool
}

func (*downloadCmd) Name() string {
	return "download"
}

func (*downloadCmd) Synopsis() string {
	return "download minecraft server binary."
}

func (*downloadCmd) Usage() string {
	return "download [-registry] [-target] [-force]\n\tDownload Minecraft Server Binary\n"
}

func (c *downloadCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.registry, "registry", "https://registry-1.docker.io/v2/", "Docker Registry Endpoint")
	f.StringVar(&c.target, "target", "bin", "Download Target")
	f.BoolVar(&c.force, "force", false, "Force Download")
}

func (c *downloadCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) (ret subcommands.ExitStatus) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\033[5;91mError: ", r)
			ret = subcommands.ExitFailure
		}
	}()
	handleDownload(c.registry, c.target, c.force)
	return subcommands.ExitSuccess
}

type unpackCmd struct {
	data string
	apk  string
}

func (*unpackCmd) Name() string {
	return "unpack"
}

func (*unpackCmd) Synopsis() string {
	return "unpack Minecraft's apk"
}

func (*unpackCmd) Usage() string {
	return "unpack [-target] [-apk]\n\tUnpack Minecraft Apk\n"
}

func (c *unpackCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.data, "target", "data", "Unpack Target")
	f.StringVar(&c.apk, "apk", "minecraft.apk", "Unpack Source")
}

func (c *unpackCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) (ret subcommands.ExitStatus) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\033[5;91mError: ", r)
			ret = subcommands.ExitFailure
		}
	}()
	unpack(c.data, c.apk)
	return subcommands.ExitSuccess
}

type runCmd struct {
	bin  string
	data string
	link string
}

func (*runCmd) Name() string {
	return "run"
}

func (*runCmd) Synopsis() string {
	return "run minecraft server"
}

func (*runCmd) Usage() string {
	return "run [-bin] [-data] [-link]\n\tRun Minecraft Server\n"
}

func (c *runCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.data, "data", "data", "Minecraft Data Directory")
	f.StringVar(&c.bin, "bin", "bin", "Minecraft Server Binary Path")
	f.StringVar(&c.link, "link", "games", "World Link Path")
}

func prepare(data string, link string) {
	games := filepath.Join(data, "games")
	props := filepath.Join(data, "server.properties")
	linkProps := filepath.Join(link, "server.properties")
	gamesProps := filepath.Join(games, "server.properties")
	if _, err := os.Stat(link); os.IsNotExist(err) {
		os.MkdirAll(link, os.ModePerm)
	}
	if _, err := os.Stat(linkProps); os.IsNotExist(err) {
		f, err := os.OpenFile(linkProps, os.O_RDONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			panic(err)
		}
		if err = f.Close(); err != nil {
			panic(err)
		}
	}
	os.RemoveAll(games)
	os.RemoveAll(props)
	os.Symlink(link, games)
	os.Symlink(gamesProps, props)
}

func (c *runCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) (ret subcommands.ExitStatus) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\033[5;91mError: \n", r)
			ret = subcommands.ExitFailure
		}
	}()
	c.data, _ = filepath.Abs(c.data)
	c.link, _ = filepath.Abs(c.link)
	c.bin, _ = filepath.Abs(c.bin)
	prepare(c.data, c.link)
	run(c.bin, c.data)
	return subcommands.ExitSuccess
}

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&downloadCmd{}, "")
	subcommands.Register(&unpackCmd{}, "")
	subcommands.Register(&runCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
