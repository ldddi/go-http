package main

import (
	"flag"
	"httpserver/pkg/app"
	"httpserver/pkg/server"

	"net/http"

	"github.com/gorilla/mux"
)

func NewApp(name string) *app.App {
	app := &app.App{}
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	fs.StringVar(&app.ConfigFilePath, "config", "", "server config")
	fs.StringVar(&app.WorkDir, "dir", "", "server upload rootDir")
	fs.StringVar(&app.Addr, "addr", "", "address to listen")
	fs.Int64Var(&app.MaxUploadSize, "max_upload_size", 0, "max upload size in bytes")
	fs.IntVar(&app.ShutdownTimeout, "shutdown_time", 0, "shutdown timeout in milliseconds")
	fs.IntVar(&app.ReadTimeout, "read_timeout", 0, "read timeout in milliseconds")
	fs.IntVar(&app.WriteTimeout, "write_time", 0, "write timeout in milliseconds")
	fs.BoolVar(&app.EnableAuth, "enable_auth", false, "read timeout. zero or negative value means no timeout. can be suffixed by the time units. If no suffix is provided, it is interpreted as seconds.")
	fs.BoolVar(&app.EnableCORS, "enable_cors", false, "write timeout. zero or negative value means no timeout. same format as read_timeout.")
	app.FlagSet = fs
	return app
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{path:.*}", server.BrowserGetHandler)

	http.ListenAndServe(":8888", r)
}
