package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	fileUtils "github.com/nvthongswansea/xtreme/internal/file-utils"
	restful "github.com/nvthongswansea/xtreme/internal/fman/delivery/restful"
	_fmanRepo "github.com/nvthongswansea/xtreme/internal/fman/repo"
	_fmanUC "github.com/nvthongswansea/xtreme/internal/fman/usecase"
	uuidUtils "github.com/nvthongswansea/xtreme/internal/uuid-utils"
)

var basePath string

func init() {
	// Get basePath from cmd args
	flag.StringVar(&basePath, "base_path", "", "Base path that stores all files")
	flag.Parse()
	if basePath == "" {
		fmt.Println("base_path arg is missing!")
		os.Exit(1)
	}
}

func main() {
	sqliteRepo := _fmanRepo.NewFManSQLiteRepo()
	uuidGenerator := &uuidUtils.GoogleUUIDGenerator{}
	fileSaver := fileUtils.CreateNewLocalFileOperator(basePath)
	fmanUC := _fmanUC.NewFManLocalUsecase(sqliteRepo, uuidGenerator, fileSaver)
	//Start web service
	e := echo.New()
	restful.InitFmanHandler(e, fmanUC)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.Fatal(e.Start(":4000"))
}
