package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"httpserver/pkg/server"
	"httpserver/pkg/utils"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type boolPrtFlag struct {
	val   bool
	isSet bool
}

type boolOpt boolPrtFlag

func (b *boolOpt) String() string {
	if b == nil {
		return ""
	}
	return fmt.Sprintf("%v", b.Val())
}

func (b *boolOpt) Set(value string) error {
	v, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}

	b.val = v
	b.isSet = true
	return nil
}

// support --enable_auth and --no-enable_auth
func (b *boolOpt) IsBoolFlag() bool {
	return true
}

func (b *boolOpt) IsSet() bool {
	return b.isSet
}

func (b *boolOpt) Val() bool {
	if b == nil {
		return false
	}

	return b.val
}

func BoolPointer(v bool) *bool {
	return &v
}

// args config
type App struct {
	FlagSet        *flag.FlagSet
	Addr           string
	WorkDir        string
	ConfigFilePath string

	EnableAuth boolOpt
	EnableCORS boolOpt

	FileNamingStrategy string
}

func (a *App) String() string {
	enableAuthStr := "<nil>"
	if a.EnableAuth.IsSet() {
		enableAuthStr = fmt.Sprint(a.EnableAuth.Val())
	}
	enableCORSStr := "<nil>"
	if a.EnableCORS.IsSet() {
		enableCORSStr = fmt.Sprint(a.EnableCORS.Val())
	}
	return fmt.Sprintf(
		"App{\n"+
			"\tFlagSet: %p\n"+
			"\tAddr: %q\n"+
			"\tWorkDir: %q\n"+
			"\tConfigFilePath: %q\n"+
			"\tEnableAuth: %s\n"+
			"\tEnableCORS: %s\n"+
			"\tFileNamingStrategy: %q\n"+
			"}",
		a.FlagSet,
		a.Addr,
		a.WorkDir,
		a.ConfigFilePath,
		enableAuthStr,
		enableCORSStr,
		a.FileNamingStrategy,
	)
}

func (a *App) Run(args []string) {
	a.ParseConfig(args)
}

func (a *App) ParseConfig(args []string) {
	if err := a.FlagSet.Parse(args); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v\n", a)

	// default config file
	if a.ConfigFilePath == "" {
		rootDir, err := utils.GetProjectRoot()
		if err != nil {
			log.Fatal(err)
		}
		a.ConfigFilePath = filepath.Join(rootDir, "config.json")
	}

	data, err := os.ReadFile(a.ConfigFilePath)
	if err != nil {
		log.Fatal(err)
	}
	serverConfig := &server.ServerConfig{}
	json.Unmarshal(data, serverConfig)

	fmt.Println(serverConfig)
}
