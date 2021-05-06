package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	restful "github.com/nvthongswansea/xtreme/internal/fman/delivery/restful"
	_fmanRepo "github.com/nvthongswansea/xtreme/internal/fman/repo"
	_fmanUC "github.com/nvthongswansea/xtreme/internal/fman/usecase"
	fileUtils "github.com/nvthongswansea/xtreme/pkg/file-utils"
	uuidUtils "github.com/nvthongswansea/xtreme/pkg/uuid-utils"
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
	sqliteRepo := _fmanRepo.NewFManSQLiteRepo()
	uuidGenerator := &uuidUtils.GoogleUUIDGenerator{}
	localFileOps := fileUtils.CreateNewLocalFileOperator(xtremeCfg.Backend.UploadDir)
	fmanUC := _fmanUC.NewFManLocalUsecase(sqliteRepo, sqliteRepo, sqliteRepo, uuidGenerator, localFileOps)
	//Start web service
	e := echo.New()
	restful.InitFmanHandler(e, fmanUC)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	backendAddr := fmt.Sprintf("%s:%d", xtremeCfg.Backend.Host, xtremeCfg.Backend.Port)
	e.Logger.Fatal(e.Start(backendAddr))
}
