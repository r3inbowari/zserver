package main

import (
	"github.com/r3inbowari/zlog"
	"zserver"
)

func main() {
	l := zlog.NewLogger()
	l.SetScreen(true)
	d := zserver.DefaultServer(zserver.Options{CaCert: "server.crt", CaKey: "server.key", Log: &l.Logger})
	d.Start()
}
