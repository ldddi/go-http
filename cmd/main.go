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

	// Custom struct to distinguish whether user explicitly sets true or false.
	fs.Var(&app.EnableAuth, "enable_auth", "read timeout. zero or negative value means no timeout. can be suffixed by the time units. If no suffix is provided, it is interpreted as seconds.")
	fs.Var(&app.EnableCORS, "enable_cors", "write timeout. zero or negative value means no timeout. same format as read_timeout.")

	fs.StringVar(&app.FileNamingStrategy, "file_naming_strategy", "", "file naming strategy, default is uuid, can be uuid or original")
	app.FlagSet = fs
	return app
}
func main() {
	app := NewApp(os.Args[0])

	app.Run(os.Args[1:])
	// r := mux.NewRouter()
	// r.HandleFunc("/{path:.*}", server.BrowserGetHandler)

	// http.ListenAndServe(":8888", r)
}
