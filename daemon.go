package main

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/coreos/go-systemd/daemon"
	"github.com/coreos/go-systemd/journal"
	"github.com/godbus/dbus"
)

var priMap = []journal.Priority{
	journal.PriDebug,
	journal.PriDebug,
	journal.PriInfo,
	journal.PriNotice,
	journal.PriWarning,
	journal.PriErr,
	journal.PriEmerg,
}

func runDaemon(datapath, profile string) {
	conn, err := dbus.SessionBus()
	if err != nil {
		printWarn("Failed to connect to session bus:" + err.Error())
		os.Exit(1)
	}
	defer conn.Close()
	dbusLog := make(chan *dbus.Signal, 10)
	conn.Signal(dbusLog)

	proc := make(chan bool, 1)
	f, stop := runImpl(datapath, proc)
	defer f.Close()
	defer stop()

	go func() {
		for {
			_, err := io.CopyN(ioutil.Discard, f, 1)
			if err != nil {
				return
			}
		}
	}()

	daemon.SdNotify(false, daemon.SdNotifyReady)

	for {
		select {
		case <-proc:
			break
		case v := <-dbusLog:
			if v.Name == "bedrockserver.core.log" {
				journal.Print(priMap[v.Body[0].(uint8)], "[%s] %s", v.Body[1], v.Body[2])
			}
		}
	}
}
