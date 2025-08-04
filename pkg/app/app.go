package app

import "flag"

type App struct {
	FlagSet         *flag.FlagSet
	ConfigFilePath  string
	WorkDir         string
	Addr            string
	MaxUploadSize   int64
	ShutdownTimeout int
	ReadTimeout     int
	WriteTimeout    int

	EnableAuth bool
	EnableCORS bool

	FileNamingStrategy string
}

func (a *App) Run(args []string) {
	a.ParseConfig(args)
}

func (a *App) ParseConfig(args []string) {

}
