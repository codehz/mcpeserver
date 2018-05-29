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
	bin     string
	data    string
	link    string
	prompt  string
	logfile string
}

func (*runCmd) Name() string {
	return "run"
}

func (*runCmd) Synopsis() string {
	return "run minecraft server"
}

func (*runCmd) Usage() string {
	return "run [-bin] [-data] [-link] [-prompt] [-websocket] [-token] [-logfile]\n\tRun Minecraft Server\n"
}

func (c *runCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.data, "data", "data", "Minecraft Data Directory")
	f.StringVar(&c.bin, "bin", "bin", "Minecraft Server Binary Path")
	f.StringVar(&c.link, "link", "games", "World Link Path")
	f.StringVar(&c.prompt, "prompt", "{{esc}}[0;36;1mmcpe:{{esc}}[22m//{{username}}@{{hostname}}$ {{esc}}[33;4m", "Prompt String Template")
	f.StringVar(&c.logfile, "logfile", "games/mcpeserver.log", "Log File Path")
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
	for run(c.bin, c.data, c.logfile, fasttemplate.New(c.prompt, "{{", "}}")) {
		printInfo("restarting...")
	}
	return subcommands.ExitSuccess
}

type modsCmd struct {
	endpoint string
	info     string
	remote   bool
	download string
}

func (*modsCmd) Name() string     { return "mods" }
func (*modsCmd) Synopsis() string { return "Mods Management" }
func (*modsCmd) Usage() string    { return "mods [--endpoint] [--info] [--remote] [--download]\n" }
func (c *modsCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.endpoint, "endpoint", "https://mcpe.codehz.one/", "Mods Repo Endpoint")
	f.StringVar(&c.info, "info", "", "Display a Remote Mods' info")
	f.BoolVar(&c.remote, "remote", false, "List Remote Mods")
	f.StringVar(&c.download, "download", "", "Download Mod")
}
func (c *modsCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) (ret subcommands.ExitStatus) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\033[5;91mError: \n", r)
			ret = subcommands.ExitFailure
		}
	}()
	if c.remote {
		listRemoteMod(c.endpoint)
	} else if len(c.info) > 0 {
		infoRemoteMod(c.endpoint, c.info)
	} else if len(c.download) > 0 {
		downloadMod(c.endpoint, c.download)
	} else {
		listLocalMod()
	}
	return subcommands.ExitSuccess
}

type daemonCmd struct {
	bin     string
	data    string
	link    string
	logfile string
	socket  string
}

func (*daemonCmd) Name() string     { return "daemon" }
func (*daemonCmd) Synopsis() string { return "Daemon" }
func (*daemonCmd) Usage() string {
	return "daemon [-bin] [-data] [-link] [-daemon]\n\tRun server as daemon"
}
func (d *daemonCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&d.data, "data", "data", "Minecraft Data Directory")
	f.StringVar(&d.bin, "bin", "bin", "Minecraft Server Binary Path")
	f.StringVar(&d.link, "link", "games", "World Link Path")
	f.StringVar(&d.logfile, "logfile", "games/mcpeserver.log", "Log File Path")
	f.StringVar(&d.socket, "socket", "games/mcpeserver.sock", "Socket File Path")
}
func (d *daemonCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) (ret subcommands.ExitStatus) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\033[5;91mError: \n", r)
			ret = subcommands.ExitFailure
		}
	}()
	d.data, _ = filepath.Abs(d.data)
	d.link, _ = filepath.Abs(d.link)
	d.bin, _ = filepath.Abs(d.bin)
	prepare(d.data, d.link)
	runDaemon(d.bin, d.data, d.logfile, d.socket)
	return subcommands.ExitSuccess
}

type updateCmd struct {
	path string
}

func (*updateCmd) Name() string {
	return "update"
}

func (*updateCmd) Synopsis() string {
	return "Update Self"
}

func (*updateCmd) Usage() string {
	return "update [-target]\n"
}

func (c *updateCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.path, "path", "./mcpeserver", "Download target")
}

func (c *updateCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) (ret subcommands.ExitStatus) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\033[5;91mError: \n", r)
			ret = subcommands.ExitFailure
		}
	}()
	printInfo("Get Latest Release...")
	url := getServerURL()
	printPair("URL", url)
	fetchBinary(url, c.path)
	return subcommands.ExitSuccess
}

type versionCmd struct{}

func (*versionCmd) Name() string             { return "version" }
func (*versionCmd) Synopsis() string         { return "Show version" }
func (*versionCmd) Usage() string            { return "version" }
func (*versionCmd) SetFlags(f *flag.FlagSet) {}
func (*versionCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) (ret subcommands.ExitStatus) {
	printPair("Version", VERSION)
	return subcommands.ExitSuccess
}

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&downloadCmd{}, "")
	subcommands.Register(&unpackCmd{}, "")
	subcommands.Register(&runCmd{}, "")
	subcommands.Register(&daemonCmd{}, "")
	subcommands.Register(&updateCmd{}, "")
	subcommands.Register(&modsCmd{}, "")
	subcommands.Register(&versionCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
