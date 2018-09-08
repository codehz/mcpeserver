package main

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/chzyer/readline"
	"github.com/valyala/fasttemplate"
)

func attach(profile string, prompt *fasttemplate.Template) {
	var bus bus
	bus.init(profile)
	defer bus.close()

	v, err := bus.ping()
	if err != nil {
		printWarn("Service is not running")
		os.Exit(1)
	}
	printPair("Service Version", v)

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
	go func() {
		for v := range bus.log {
			if v.Name == "one.codehz.bedrockserver.core.log" {
				fmt.Fprintf(lw, "\033[0m%s [%v] %v\033[0m\n", table[v.Body[0].(uint8)], v.Body[1], v.Body[2])
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
		result, err := bus.exec(ncmd)
		if err != nil {
			fmt.Fprintf(lw, "\033[0m%v\033[0m\n", err)
		} else {
			if len(result) > 0 {
				fmt.Fprintf(lw, "\033[0m%s\n\033[0m", replacer.Replace(result))
			}
		}
	}
}
