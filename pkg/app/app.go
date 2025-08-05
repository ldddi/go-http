package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"httpserver/pkg/server"
	"httpserver/pkg/utils"
	"log"
	"os"
	"strconv"
	"time"

	"dario.cat/mergo"
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

var DefaultConfig = server.ServerConfig{
	Addr:               "127.0.0.1:8080",
	WorkDir:            "",
	FileNamingStrategy: "uuid",

	MaxUploadSize:   1024 * 1024,
	ShutdownTimeout: 15000,
	ReadTimeout:     time.Duration(15 * time.Second),
	WriteTimeout:    0,

	EnableCORS: nil,
	EnableAuth: nil,
}

// args config
type App struct {
	FlagSet *flag.FlagSet

	Addr               string
	WorkDir            string
	ConfigFilePath     string
	FileNamingStrategy string

	MaxUploadSize   int64
	ShutdownTimeout int
	ReadTimeout     int
	WriteTimeout    int

	EnableAuth boolOpt
	EnableCORS boolOpt
}

func (a *App) Run(args []string) {
	config, err := a.ParseConfig(args)
	if err != nil {
		log.Fatal(err)
	}

	server.NewServer(*config)
}

func (a *App) ParseConfig(args []string) (*server.ServerConfig, error) {
	if err := a.FlagSet.Parse(args); err != nil {
		return nil, fmt.Errorf("failed parse flags: %w", err)
	}

	config := DefaultConfig
	log.Printf("default config: %+v\n", config)

	if a.ConfigFilePath != "" {
		f, err := os.Open(a.ConfigFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed open configFile: %w", err)
		}
		defer f.Close()

		fileConfig := server.ServerConfig{}
		if err := json.NewDecoder(f).Decode(&fileConfig); err != nil {
			return nil, fmt.Errorf("failed decode config file: %w", err)
		}

		if err := mergo.Merge(&config, fileConfig, mergo.WithOverride); err != nil {
			return nil, fmt.Errorf("failed merge default and fileconfig: %w", err)
		}
		log.Printf("default config and fileconfig merge result: %+v\n", config)
	} else {
		log.Printf("no provided fileconfig\n")
	}

	if a.WorkDir == "" {
		rootDir, err := utils.GetProjectRoot()
		if err != nil {
			return nil, err
		}

		a.WorkDir = rootDir
		log.Printf("no provided WorkDir, use default WorkDir in \"%v\"", a.WorkDir)
	} else {
		log.Printf("use WorkDir in \"%v\"", a.WorkDir)
	}

	argsConfig := server.ServerConfig{
		Addr:               a.Addr,
		WorkDir:            a.WorkDir,
		FileNamingStrategy: a.FileNamingStrategy,

		MaxUploadSize:   a.MaxUploadSize,
		ShutdownTimeout: a.ShutdownTimeout,
		ReadTimeout:     time.Duration(a.ReadTimeout),
		WriteTimeout:    time.Duration(a.WriteTimeout),
	}

	if a.EnableCORS.isSet {
		argsConfig.EnableCORS = &a.EnableCORS.val
	}
	if a.EnableAuth.isSet {
		argsConfig.EnableAuth = &a.EnableAuth.val
	}

	if err := mergo.Merge(&config, argsConfig, mergo.WithOverride); err != nil {
		return nil, fmt.Errorf("failed to merge config from flags: %w", err)
	}
	log.Printf("final config: %+v\n", config)

	return &config, nil
}
