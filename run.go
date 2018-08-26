package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/chzyer/readline"
	"github.com/godbus/dbus"
	"github.com/kr/pty"
	"github.com/valyala/fasttemplate"
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
		output(strings.TrimRight(line, "\n"))
	}
}

func runImpl(done chan bool) (*os.File, func()) {
	cmd := exec.Command("./bin/bedrockserver")
	cmd.Dir, _ = os.Getwd()
	f, err := pty.Start(cmd)
	if err != nil {
		panic(err)
	}
	status := true
	selfLock := make(chan struct{}, 1)
	go func() {
		cmd.Wait()
		selfLock <- struct{}{}
		done <- status
	}()
	return f, func() {
		status = false
		cmd.Process.Signal(os.Interrupt)
		<-selfLock
	}
}

var table = []string{"T", "D", "I", "N", "W", "E", "F"}

func run(profile string, prompt *fasttemplate.Template) bool {
	conn, err := dbus.SystemBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		os.Exit(1)
	}
	defer conn.Close()
	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		fmt.Sprintf("type='signal',path='/',interface='bedrockserver.core',sender='one.codehz.bedrockserver.%s'", profile))
	dbusLog := make(chan *dbus.Signal, 10)
	conn.Signal(dbusLog)
	dbusObj := conn.Object("one.codehz.bedrockserver."+profile, "/one/codehz/bedrockserver")

	log, err := os.OpenFile(profile+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		printWarn("Log File load failed")
		return false
	}
	defer log.Close()
	proc := make(chan bool, 1)
	f, stop := runImpl(proc)
	defer f.Close()
	defer stop()
	username := "nobody"
	hostname := "bedrockserver"
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
	rl, _ := readline.NewEx(&readline.Config{
		Prompt: prompt.ExecuteString(map[string]interface{}{
			"username": username,
			"hostname": hostname,
			"esc":      "\033",
		}),
		HistoryFile:     ".readline-history",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       ":quit",

		HistorySearchFold: true,
		FuncFilterInputRune: func(r rune) (rune, bool) {
			if r == readline.CharCtrlZ {
				return r, false
			}
			return r, true
		},
	})
	defer rl.Close()
	lw := io.MultiWriter(rl.Stdout(), log)
	status := false
	queue := make(map[uint32]bool)
	execFn := func(src, cmd string) {
		ncmd := strings.TrimSpace(cmd)
		if len(ncmd) == 0 {
			return
		}
		fmt.Fprintf(log, "%s>%s\n", src, ncmd)
		switch {
		case strings.HasPrefix(ncmd, ":restart"):
			status = true
			rl.Close()
		case strings.HasPrefix(ncmd, ":quit"):
			status = true
			rl.Close()
		default:
			var rid uint32
			err := dbusObj.Call("bedrockserver.core.exec", 0, ncmd).Store(&rid)
			if err != nil {
				fmt.Fprintf(lw, "\033[0m%v\033[0m\n", err)
			} else {
				queue[rid] = true
			}
		}
	}
	go packOutput(f, func(text string) {
		fmt.Fprintf(lw, "\033[0m%s\033[0m\n", text)
	})
	go func() {
		for v := range dbusLog {
			if v.Name == "bedrockserver.core.log" {
				fmt.Fprintf(lw, "\033[0m%s [%v] %v\033[0m\n", table[v.Body[0].(uint8)], v.Body[1], v.Body[2])
			} else if v.Name == "bedrockserver.core.exec_result" {
				if _, ok := queue[v.Body[0].(uint32)]; ok {
					fmt.Fprintf(lw, "\033[0m%s\n\033[0m", replacer.Replace(v.Body[1].(string)))
				}
			}
		}
	}()
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
		case strings.HasPrefix(line, ":restart"):
			return true
		case strings.HasPrefix(line, ":quit"):
			return status
		default:
			execFn("console", line)
		}
	}
	return false
}
