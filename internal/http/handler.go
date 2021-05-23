package http

import (
	"github.com/labstack/echo/v4"
	"github.com/nvthongswansea/xtreme/internal/authen"
	"github.com/nvthongswansea/xtreme/internal/author"
	"github.com/nvthongswansea/xtreme/internal/ent"
	"github.com/nvthongswansea/xtreme/internal/file-manager/local"
	"github.com/nvthongswansea/xtreme/internal/repository/directory"
	"github.com/nvthongswansea/xtreme/internal/repository/file"
	"github.com/nvthongswansea/xtreme/internal/repository/role"
	"github.com/nvthongswansea/xtreme/internal/repository/transaction"
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
	fileRepo := file.NewEntSQLFileRepo(client, uuidUtils)
	dirRepo := directory.NewEntSQLDirectoryRepo(client, uuidUtils)
	txRepo := transaction.NewEntSQLTxRepo(client)
	fileOps := fileUtils.CreateNewLocalFileOperator(basePath)
	fileCompress := fileUtils.CreateNewFileZipper(basePath, "")

	authService := authen.NewLocalAuthenticator(userRepo, txRepo, dirRepo, passwordUtils, "test")
	attachAuthenHTTPHandlerHandler(e, authService)

	roleRepo := role.NewEntSQLRoleRepo(client, uuidUtils)
	authorRepo, err := author.NewCasbinAuthorizer("casbin-cfg/model.conf", "casbin-cfg/policy.csv", roleRepo)
	if err != nil {
		panic(err)
	}
	localFManService := local.NewMultiOSFileManager(fileRepo, dirRepo, txRepo, uuidUtils, fileOps, fileCompress, authorRepo)
	attachLocalFManHTTPHandler(e, localFManService, authService)
}
