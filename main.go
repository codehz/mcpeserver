package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/valyala/fasttemplate"
)

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

type attachCmd struct {
	profile string
	prompt  string
}

func (*attachCmd) Name() string     { return "attach" }
func (*attachCmd) Synopsis() string { return "attach daemon" }
func (*attachCmd) Usage() string    { return "attach [-profile] [-prompt]" }
func (a *attachCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&a.profile, "profile", "default", "Game Profile")
	f.StringVar(&a.prompt, "prompt", "{{esc}}[0;36;1msocket:{{esc}}[22m//{{username}}@{{hostname}}$ {{esc}}[33;4m", "Prompt String Template")
}
func (a *attachCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) (ret subcommands.ExitStatus) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\033[5;91mError: \n", r)
			ret = subcommands.ExitFailure
		}
	}()
	attach(a.profile, fasttemplate.New(a.prompt, "{{", "}}"))
	return subcommands.ExitSuccess
}

type runCmd struct {
	profile string
	prompt  string
}

func (*runCmd) Name() string {
	return "run"
}

func (*runCmd) Synopsis() string {
	return "run minecraft server"
}

func (*runCmd) Usage() string {
	return "run [-profile] [-prompt] \n\tRun Minecraft Server\n"
}

func (c *runCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.profile, "profile", "default", "Game Proile")
	f.StringVar(&c.prompt, "prompt", "{{esc}}[0;36;1mmcpe:{{esc}}[22m//{{username}}@{{hostname}}$ {{esc}}[33;4m", "Prompt String Template")
}

func checkBin() {
	if _, err := os.Stat("./bin"); err != nil {
		printWarn("/bin not found, checking /opt/mcpeserver-core...")
		if _, err = os.Stat("/opt/mcpeserver-core"); err != nil {
			printWarn("/opt/mcpeserver-core not found, exiting...")
			os.Exit(1)
		} else {
			os.Symlink("/opt/mcpeserver-core", "bin")
		}
	}
}

func (c *runCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) (ret subcommands.ExitStatus) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\033[5;91mError: \n", r)
			ret = subcommands.ExitFailure
		}
	}()
	checkBin()
	for run(c.profile, fasttemplate.New(c.prompt, "{{", "}}")) {
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
	f.StringVar(&c.info, "info", "", "Display a Remote Mod' info")
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
	profile string
	systemd bool
}

func (*daemonCmd) Name() string     { return "daemon" }
func (*daemonCmd) Synopsis() string { return "Daemon" }
func (*daemonCmd) Usage() string {
	return "daemon [-profile] [-systemd]\n\tRun server as daemon"
}
func (d *daemonCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&d.profile, "profile", "default", "Game Profile")
	f.BoolVar(&d.systemd, "systemd", false, "Systemd mode")
}
func (d *daemonCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) (ret subcommands.ExitStatus) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\033[5;91mError: \n", r)
			ret = subcommands.ExitFailure
		}
	}()
	checkBin()
	runDaemon(d.profile, d.systemd)
	return subcommands.ExitSuccess
}

type updateCmd struct {
	path string
}

func (*updateCmd) Name() string {
	return "update"
}

func (*updateCmd) Synopsis() string {
	return "Self-Update"
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
	subcommands.Register(&unpackCmd{}, "")
	subcommands.Register(&attachCmd{}, "")
	subcommands.Register(&runCmd{}, "")
	subcommands.Register(&daemonCmd{}, "")
	subcommands.Register(&updateCmd{}, "")
	subcommands.Register(&modsCmd{}, "")
	subcommands.Register(&versionCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
