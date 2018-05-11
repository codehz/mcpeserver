package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/subcommands"
	"github.com/valyala/fasttemplate"
)

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
	bin       string
	data      string
	link      string
	prompt    string
	websocket string
}

func (*runCmd) Name() string {
	return "run"
}

func (*runCmd) Synopsis() string {
	return "run minecraft server"
}

func (*runCmd) Usage() string {
	return "run [-bin] [-data] [-link] [-prompt] [-websocket]\n\tRun Minecraft Server\n"
}

func (c *runCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.data, "data", "data", "Minecraft Data Directory")
	f.StringVar(&c.bin, "bin", "bin", "Minecraft Server Binary Path")
	f.StringVar(&c.link, "link", "games", "World Link Path")
	f.StringVar(&c.prompt, "prompt", "{{esc}}[0;36;1mmcpe:{{esc}}[22m//{{username}}@{{hostname}}$ {{esc}}[33;4m", "Prompt String Template")
	f.StringVar(&c.websocket, "websocket", "", "WebSocket Server Port(Disabled If Blank)")
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
	run(c.bin, c.data, c.websocket, fasttemplate.New(c.prompt, "{{", "}}"))
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
