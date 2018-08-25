package main

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/chzyer/readline"
	"github.com/godbus/dbus"
	"github.com/valyala/fasttemplate"
)

func attach(profile string, prompt *fasttemplate.Template) {
	conn, err := dbus.SessionBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		os.Exit(1)
	}
	defer conn.Close()
	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		fmt.Sprintf("type='signal',path='/',interface='bedrockserver.core',sender='one.codehz.bedrockserver.%s'", profile))
	dbusLog := make(chan *dbus.Signal, 10)
	conn.Signal(dbusLog)
	dbusObj := conn.Object("one.codehz.bedrockserver."+profile, "/")

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
	rl, _ := readline.NewEx(&readline.Config{
		Prompt: prompt.ExecuteString(map[string]interface{}{
			"username": username,
			"hostname": hostname,
			"esc":      "\033",
		}),
		HistoryFile:     ".readline-history",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       ":detach",

		HistorySearchFold: true,
		FuncFilterInputRune: func(r rune) (rune, bool) {
			if r == readline.CharCtrlZ {
				return r, false
			}
			return r, true
		},
	})
	lw := rl.Stdout()
	queue := make(map[uint32]bool)
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
		if err != nil {
			break
		}
		ncmd := strings.TrimSpace(line)
		if len(ncmd) == 0 {
			continue
		} else if ncmd == ":detach" {
			break
		} else if strings.HasPrefix(ncmd, ":") {
			fmt.Fprintln(lw, "\033[0mPlease use systemctl to control service.\033[0m")
			continue
		}
		var rid uint32
		err = dbusObj.Call("bedrockserver.core.exec", 0, ncmd).Store(&rid)
		if err != nil {
			fmt.Fprintf(lw, "\033[0m%v\033[0m\n", err)
		} else {
			queue[rid] = true
		}
	}
}
