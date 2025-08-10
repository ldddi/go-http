package main

import (
	"flag"
	"httpserver/pkg/app"
	"os"
)

func NewApp(name string) *app.App {
	app := &app.App{}
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	fs.StringVar(&app.ConfigFilePath, "config", "", "server config")
	fs.StringVar(&app.WorkDir, "dir", "", "server upload rootDir")
	fs.StringVar(&app.Addr, "addr", "", "address to listen")
	fs.Var(&app.EnableAuth, "enable_auth", "read timeout. zero or negative value means no timeout. can be suffixed by the time units. If no suffix is provided, it is interpreted as seconds.")
	app.FlagSet = fs
	return app
}
func main() {
	app := NewApp(os.Args[0])

	app.Run(os.Args[1:])
}
