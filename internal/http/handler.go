package http

import (
	"github.com/labstack/echo/v4"
	"github.com/nvthongswansea/xtreme/internal/authen"
	"github.com/nvthongswansea/xtreme/internal/author"
	"github.com/nvthongswansea/xtreme/internal/ent"
	"github.com/nvthongswansea/xtreme/internal/file-manager/local"
	"github.com/nvthongswansea/xtreme/internal/repository/directory"
	"github.com/nvthongswansea/xtreme/internal/repository/file"
	"github.com/nvthongswansea/xtreme/internal/repository/user"
	"github.com/nvthongswansea/xtreme/pkg/fileUtils"
	"github.com/nvthongswansea/xtreme/pkg/pwd"
	uuidUtils "github.com/nvthongswansea/xtreme/pkg/uuidUtils"
)

// InitHTTPHandler
func InitHTTPHandler(e *echo.Echo, client *ent.Client, basePath string) {
	uuidUtils := uuidUtils.GoogleUUIDGenerator{}
	passwordUtils := pwd.NewBCryptHashComparer(10)
	userRepo := user.NewEntSQLUserRepo(client, uuidUtils)

	authService := authen.NewLocalAuthenticator(userRepo, passwordUtils, "test")
	attachAuthenHTTPHandlerHandler(e, authService)

	fileRepo := file.NewEntSQLFileRepo(client, uuidUtils)
	dirRepo := directory.NewEntSQLDirectoryRepo(client, uuidUtils)
	fileOps := fileUtils.CreateNewLocalFileOperator(basePath)
	fileCompress := fileUtils.CreateNewFileZipper(basePath, "")
	localFManService := local.NewMultiOSFileManager(fileRepo, dirRepo, uuidUtils, fileOps, fileCompress, author.StubAuthorizer{})
	attachLocalFManHTTPHandler(e, localFManService, authService)
}
