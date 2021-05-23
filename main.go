package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nvthongswansea/xtreme/internal/ent"
	"github.com/nvthongswansea/xtreme/internal/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

var configFilePath string
var xtremeCfg *Config

func init() {
	// Get config file path from cmd args
	flag.StringVar(&configFilePath, "config_file", "", "Path of the config file")
	flag.Parse()
	if configFilePath == "" {
		fmt.Println("config_file arg is missing!")
		os.Exit(1)
	}
	// Get config from config file
	var err error
	xtremeCfg, err = NewConfig(configFilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Set log level
	switch xtremeCfg.LogLevel {
	case "panic":
		log.SetLevel(log.PanicLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}
}

func main() {
	client, err := ent.Open("sqlite3", "app.db:ent?mode=memory&cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	e := echo.New()
	http.InitHTTPHandler(e, client, xtremeCfg.Backend.UploadDir)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//e.Logger = &log.Logger{}
	backendAddr := fmt.Sprintf("%s:%d", xtremeCfg.Backend.Host, xtremeCfg.Backend.Port)
	e.Logger.Fatal(e.Start(backendAddr))
}
