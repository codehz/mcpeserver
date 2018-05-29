package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func runDaemon(base, datapath, logfile, socket string) {
	log, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer log.Close()
	fmt.Fprintln(log, "Starting...")
	defer fmt.Fprintln(log, "Stopping...")
	sigs := make(chan os.Signal, 1)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	l, err := net.Listen("unix", socket)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	defer os.Remove(socket)
	exec := make(chan string)
	defer close(exec)
	conns := make(map[net.Conn]bool)
	newConns := make(chan net.Conn, 128)
	deadConns := make(chan net.Conn, 128)
	publishes := make(chan string, 128)
	done := make(chan struct{}, 1)
	defer func() {
		done <- struct{}{}
	}()
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				break
			}
			newConns <- c
		}
	}()
	go func() {
		defer close(newConns)
		defer close(deadConns)
		defer close(publishes)
		defer close(done)
	outer:
		for {
			select {
			case c := <-newConns:
				conns[c] = true
				go func() {
					bs := bufio.NewScanner(c)
					for bs.Scan() {
						exec <- bs.Text()
					}
					deadConns <- c
				}()
			case c := <-deadConns:
				_ = c.Close()
				delete(conns, c)
			case text := <-publishes:
				for conn := range conns {
					go fmt.Fprintln(conn, text)
				}
			case <-done:
				break outer
			}
		}
	}()
	doLog := func(text string) {
		fmt.Fprintf(log, "%s\n", text)
		publishes <- text
	}
	for func() bool {
		f, quit := runImpl(base, datapath)
		defer f.Close()
		defer quit()
		status := make(chan bool)
		defer close(status)
		execFn := func(src, cmd string) {
			fmt.Fprintf(f, "%s\n", cmd)
			doLog(fmt.Sprintf("%s>%s", src, cmd))
			switch {
			case strings.HasPrefix(cmd, ":restart"):
				status <- true
			case strings.HasPrefix(cmd, ":quit"):
				status <- false
			}
		}
		cache := 0
		go packOutput(f, func(text string) {
			if strings.HasPrefix(text, "\x07") {
				execFn("mod", text[1:len(text)-1])
				cache++
			} else {
				if cache == 0 {
					doLog(fmt.Sprintf("\033[0m%s\033[0m", replacer.Replace(text)))
				} else {
					cache--
				}
			}
		})
		for {
			select {
			case x := <-status:
				return x
			case <-sigs:
				return false
			case line, ok := <-exec:
				if ok {
					cache++
					go execFn("socket", line)
				} else {
					return false
				}
			}
		}
	}() {
		doLog("Restarting...")
	}
	return
}
