package main

import (
	"fmt"
	"os"

	"github.com/godbus/dbus"
)

type bus struct {
	conn *dbus.Conn
	log  chan *dbus.Signal
	obj  dbus.BusObject
}

func (b *bus) init(profile string) {
	var err error
	b.conn, err = dbus.SystemBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		os.Exit(1)
	}
	b.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		fmt.Sprintf("type='signal',path='/one/codehz/bedrockserver',interface='one.codehz.bedrockserver.core',sender='one.codehz.bedrockserver.%s'", profile))
	b.log = make(chan *dbus.Signal, 10)
	b.conn.Signal(b.log)
	b.obj = b.conn.Object("one.codehz.bedrockserver."+profile, "/one/codehz/bedrockserver")
}

func (b bus) close() {
	b.conn.Close()
}

func (b bus) exec(cmd string) (uint32, error) {
	var rid uint32
	err := b.obj.Call("one.codehz.bedrockserver.core.exec", 0, cmd).Store(&rid)
	return rid, err
}
