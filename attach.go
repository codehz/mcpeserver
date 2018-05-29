package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/valyala/fasttemplate"
)

func attach(socket string, prompt *fasttemplate.Template) {
	c, err := net.Dial("unix", socket)
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
	rl, _ := readline.NewEx(&readline.Config{
		Prompt: prompt.ExecuteString(map[string]interface{}{
			"username": username,
			"hostname": hostname,
			"esc":      "\033",
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
	// lw := rl.Stdout()
	cache := 0
	buffer := make(chan string, 512)
	go func() {
		bs := bufio.NewScanner(c)
		for bs.Scan() {
			if cache == 0 {
				buffer <- bs.Text()
			} else {
				cache--
			}
		}
		rl.Close()
	}()
	go func() {
		for {
			line, ok := <-buffer
			if !ok {
				break
			}
			fmt.Fprintln(rl, line)
			time.Sleep(time.Millisecond * 10)
		}
	}()
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		fmt.Fprintln(c, line)
		cache++
	}
}
