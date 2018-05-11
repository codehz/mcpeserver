package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
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
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/valyala/fasttemplate"

	"github.com/chzyer/readline"
	"github.com/google/subcommands"
	"github.com/kr/pty"
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
	fmt.Printf("\033[0;32m%s\033[0m\n", item)
}

func printWarn(item string) {
	fmt.Printf("\033[0;91m%s\033[0m\n", item)
}

func printPair(key string, value string) {
	fmt.Printf("\033[0;34m%s: \033[0;35m%s\033[0m\n", key, value)
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

var completer = readline.NewPrefixCompleter(
	readline.PcItem("op"),
	readline.PcItem("help"),
	readline.PcItem("give"),
	readline.PcItem("always"),
	readline.PcItem("kill"),
	readline.PcItem("clear"),
	readline.PcItem("clone"),
	readline.PcItem("deop"),
	readline.PcItem("effect"),
	readline.PcItem("enchant"),
	readline.PcItem("effect"),
	readline.PcItem("execute"),
	readline.PcItem("fill"),
	readline.PcItem("gamerule",
		readline.PcItem("commandBlockOutput"),
		readline.PcItem("doDaylightCycle"),
		readline.PcItem("doEntityDrops"),
		readline.PcItem("doFireTick"),
		readline.PcItem("doMobLoot"),
		readline.PcItem("doMobSpawning"),
		readline.PcItem("doTileDrops"),
		readline.PcItem("doWeatherCycle"),
		readline.PcItem("drowningdamage"),
		readline.PcItem("falldamage"),
		readline.PcItem("firedamage"),
		readline.PcItem("keepInventory"),
		readline.PcItem("mobGriefing"),
		readline.PcItem("naturalRegeneration"),
		readline.PcItem("pvp"),
		readline.PcItem("sendCommandFeedback"),
		readline.PcItem("showcoordinates"),
		readline.PcItem("tntexplodes"),
	),
	readline.PcItem("list"),
	readline.PcItem("playsound"),
	readline.PcItem("replaceitem",
		readline.PcItem("block"),
		readline.PcItem("entity"),
	),
	readline.PcItem("setmaxplayers"),
	readline.PcItem("setworldspawn"),
	readline.PcItem("spawnpoint"),
	readline.PcItem("spreadplayers"),
	readline.PcItem("stopsound"),
	readline.PcItem("summon",
		readline.PcItem("item"),
		readline.PcItem("xp_orb"),
		readline.PcItem("tnt"),
		readline.PcItem("falling_block"),
		readline.PcItem("moving_block"),
		readline.PcItem("armor_stand"),
		readline.PcItem("xp_bottle"),
		readline.PcItem("eye_of_ender_signal"),
		readline.PcItem("ender_crystal"),
		readline.PcItem("fireworks_rocket"),
		readline.PcItem("shulker_bullet"),
		readline.PcItem("fishing_hook"),
		readline.PcItem("dragon_fireball"),
		readline.PcItem("arrow"),
		readline.PcItem("snowball"),
		readline.PcItem("egg"),
		readline.PcItem("painting"),
		readline.PcItem("minecart"),
		readline.PcItem("large_fireball"),
		readline.PcItem("splash_potion"),
		readline.PcItem("ender_pearl"),
		readline.PcItem("leash_knot"),
		readline.PcItem("wither_skull"),
		readline.PcItem("boat"),
		readline.PcItem("wither_skull_dangerous"),
		readline.PcItem("lightning_bolt"),
		readline.PcItem("small_fireball"),
		readline.PcItem("area_effect_cloud"),
		readline.PcItem("hopper_minecart"),
		readline.PcItem("tnt_minecart"),
		readline.PcItem("chest_minecart"),
		readline.PcItem("command_block_minecart"),
		readline.PcItem("lingering_potion"),
		readline.PcItem("llama_spit"),
		readline.PcItem("evocation_fang"),
		readline.PcItem("zombie"),
		readline.PcItem("creeper"),
		readline.PcItem("skeleton"),
		readline.PcItem("spider"),
		readline.PcItem("zombie_pigman"),
		readline.PcItem("slime"),
		readline.PcItem("enderman"),
		readline.PcItem("silverfish"),
		readline.PcItem("cave_spider"),
		readline.PcItem("ghast"),
		readline.PcItem("magma_cube"),
		readline.PcItem("blaze"),
		readline.PcItem("zombie_villager"),
		readline.PcItem("witch"),
		readline.PcItem("stray"),
		readline.PcItem("husk"),
		readline.PcItem("wither_skeleton"),
		readline.PcItem("guardian"),
		readline.PcItem("elder_guardian"),
		readline.PcItem("wither"),
		readline.PcItem("ender_dragon"),
		readline.PcItem("shulker"),
		readline.PcItem("endermite"),
		readline.PcItem("vindicator"),
		readline.PcItem("evocation_illager"),
		readline.PcItem("vex"),
		readline.PcItem("chicken"),
		readline.PcItem("cow"),
		readline.PcItem("pig"),
		readline.PcItem("sheep"),
		readline.PcItem("wolf"),
		readline.PcItem("villager"),
		readline.PcItem("mooshroom"),
		readline.PcItem("squid"),
		readline.PcItem("rabbit"),
		readline.PcItem("bat"),
		readline.PcItem("iron_golem"),
		readline.PcItem("snow_golem"),
		readline.PcItem("ocelot"),
		readline.PcItem("horse"),
		readline.PcItem("donkey"),
		readline.PcItem("mule"),
		readline.PcItem("skeleton_horse"),
		readline.PcItem("zombie_horse"),
		readline.PcItem("polar_bear"),
		readline.PcItem("llama"),
		readline.PcItem("parrot"),
		readline.PcItem("dolphin"),
		readline.PcItem("player"),
		readline.PcItem("npc"),
		readline.PcItem("learn_to_code_mascot"),
		readline.PcItem("tripod_camera"),
		readline.PcItem("chalkboard"),
	),
	readline.PcItem("teleport"),
	readline.PcItem("tell"),
	readline.PcItem("msg"),
	readline.PcItem("w"),
	readline.PcItem("testfor"),
	readline.PcItem("testforblock"),
	readline.PcItem("testforblocks"),
	readline.PcItem("tickingarea"),
	readline.PcItem("time",
		readline.PcItem("add"),
		readline.PcItem("query",
			readline.PcItem("daytime"),
			readline.PcItem("gametime"),
			readline.PcItem("day"),
		),
		readline.PcItem("set"),
	),
	readline.PcItem("title"),
	readline.PcItem("toggledownfall"),
	readline.PcItem("transferserver	"),
	readline.PcItem("wsserver"),
	readline.PcItem("xp"),
	readline.PcItem("tp"),
	readline.PcItem("weather",
		readline.PcItem("clear"),
		readline.PcItem("rain"),
		readline.PcItem("thunder"),
	),
	readline.PcItem("gamemode",
		readline.PcItem("survival"),
		readline.PcItem("creative"),
		readline.PcItem("adventure"),
	),
	readline.PcItem("difficulty",
		readline.PcItem("peaceful"),
		readline.PcItem("easy"),
		readline.PcItem("normal"),
		readline.PcItem("hard"),
	),
)

var replacer = strings.NewReplacer(
	"§0", "\033[30m", // black
	"§1", "\033[34m", // blue
	"§2", "\033[32m", // green
	"§3", "\033[36m", // aqua
	"§4", "\033[31m", // red
	"§5", "\033[35m", // purple
	"§6", "\033[33m", // gold
	"§7", "\033[37m", // gray
	"§8", "\033[90m", // dark gray
	"§9", "\033[94m", // light blue
	"§a", "\033[92m", // light green
	"§b", "\033[96m", // light aque
	"§c", "\033[91m", // light red
	"§d", "\033[95m", // light purple
	"§e", "\033[93m", // light yellow
	"§f", "\033[97m", // light white
	"§k", "\033[5m", // Obfuscated
	"§l", "\033[1m", // Bold
	"§m", "\033[2m", // Strikethrough
	"§n", "\033[4m", // Underline
	"§o", "\033[3m", // Italic
	"§r", "\033[0m", // Reset
	"[", "\033[1m[",
	"]", "]\033[22m",
	"(", "(\033[4m",
	")", "\033[24m)",
	"<", "\033[1m<",
	">", ">\033[22m",
)

func packOutput(input io.Reader, output func(string)) {
	reader := bufio.NewReader(input)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		output(strings.TrimRight(replacer.Replace(line), "\n"))
	}
}

func run(base string, datapath string, prompt *fasttemplate.Template) {
	abs, err := filepath.Abs(base)
	if err != nil {
		panic(err)
	}
	cmd := exec.Command(filepath.Join(abs, "server"))
	cmd.Env = append(os.Environ(), fmt.Sprintf("LD_LIBRARY_PATH=%s", abs))
	cmd.Dir = datapath
	f, err := pty.Start(cmd)
	if err != nil {
		panic(err)
	}
	username := "nobody"
	hostname := "mcpeserver"
	{
		u, err := user.Current()
		if err == nil {
			username = u.Username
		}
		hn, err := os.Hostname()
		if err == nil {
			hostname = hn
		}
	}
	// t := fasttemplate.New("\033[0;36;1mmcpe:\033[22m//{{username}}@{{hostname}}$ \033[33;4m", "{{", "}}")
	rl, _ := readline.NewEx(&readline.Config{
		Prompt: prompt.ExecuteString(map[string]interface{}{
			"username": username,
			"hostname": hostname,
		}),
		HistoryFile:     ".readline-history",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "quit",

		HistorySearchFold: true,
		FuncFilterInputRune: func(r rune) (rune, bool) {
			if r == readline.CharCtrlZ {
				return r, false
			}
			return r, true
		},
	})
	defer rl.Close()
	lw := rl.Stdout()
	cache := 0
	go packOutput(f, func(text string) {
		if cache == 0 {
			fmt.Fprintf(lw, "\033[0m%s\033[0m\n", text)
		} else {
			cache--
		}
	})
	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		switch {
		default:
			cache++
			fmt.Fprintf(f, "%s\n", line)
		}
	}
	cmd.Process.Signal(os.Interrupt)
	cmd.Wait()
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
	bin    string
	data   string
	link   string
	prompt string
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
	f.StringVar(&c.prompt, "prompt", "\033[0;36;1mmcpe:\033[22m//{{username}}@{{hostname}}$ \033[33;4m", "Prompt String Template")
}

func prepare(data string, link string) {
	games := filepath.Join(data, "games")
	props := filepath.Join(data, "server.properties")
	mods := filepath.Join(data, "mods")
	linkProps := filepath.Join(link, "server.properties")
	linkMods := filepath.Join(link, "mods")
	gamesProps := filepath.Join(games, "server.properties")
	gamesMods := filepath.Join(games, "mods")
	os.MkdirAll(link, os.ModePerm)
	os.MkdirAll(linkMods, os.ModePerm)
	if _, err := os.Stat(linkProps); os.IsNotExist(err) {
		f, err := os.OpenFile(linkProps, os.O_RDWR|os.O_CREATE, os.ModePerm)
		fmt.Fprintln(f, "motd=Minecraft Server\nlevel-dir=world\nlevel-name=Default Server World")
		if err != nil {
			panic(err)
		}
		if err = f.Close(); err != nil {
			panic(err)
		}
	}
	os.RemoveAll(games)
	os.Symlink(link, games)
	os.Symlink(gamesProps, props)
	os.Symlink(gamesMods, mods)
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
	run(c.bin, c.data, fasttemplate.New(c.prompt, "{{", "}}"))
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
