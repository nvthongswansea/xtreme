package http

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/nvthongswansea/xtreme/internal/authen"
	"github.com/nvthongswansea/xtreme/internal/file-manager/local"
	"github.com/nvthongswansea/xtreme/internal/models"
	"net/http"
	"strconv"
)

type LocalFManHTTPHandler struct {
	LocalFMan local.FileManagerService
	Authen    authen.Authenticator
}

func AttachLocalFManHandler(e *echo.Echo, l local.FileManagerService, a authen.Authenticator) {
	handler := &LocalFManHTTPHandler{LocalFMan: l, Authen: a}
	localGroup := e.Group("/local")
	localGroup.POST("/file", handler.createFile)
	localGroup.POST("/file/upload", handler.uploadFile)
	localGroup.GET("/file/:id", handler.getFile)
	localGroup.PATCH("/file/:id", handler.renameFile)
	localGroup.POST("/file/:id/copy", handler.copyFile)
	localGroup.PATCH("/file/:id/move", handler.moveFile)
	localGroup.GET("/file/:id/download", handler.downloadFile)
	localGroup.GET("/file/batch/download", handler.downloadFileBatch)
	localGroup.DELETE("/file/:id", handler.softRemoveFile)
	localGroup.DELETE("/file/:id/force", handler.hardRemoveFile)

	localGroup.POST("/dir", handler.createDir)
	localGroup.GET("/dir/:id", handler.getDirectory)
}

func (l *LocalFManHTTPHandler) uploadFile(c echo.Context) error {
	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		return resolveError(err, c, http.StatusInternalServerError)
	}
	src, err := file.Open()
	if err != nil {
		return resolveError(err, c, http.StatusInternalServerError)
	}
	defer src.Close()

	filename := c.FormValue("filename")
	parentUUID := c.FormValue("parent_uuid")
	ctx := c.Request().Context()
	jwtClaims, err := l.Authen.GetDataViaToken(ctx, c.Get("user"))
	if err != nil {
		return resolveError(err, c, http.StatusInternalServerError)
	}
	newFileUUID, err := l.LocalFMan.UploadFile(ctx, jwtClaims.UserUUID, filename, parentUUID, src)
	if err != nil {
		return resolveError(err, c, 0)
	}
	return c.JSON(http.StatusOK,
		models.SuccessFManResponse{
			EntityUUID: newFileUUID,
			Message:    "Uploaded file successfully",
		},
	)
}

func (l *LocalFManHTTPHandler) createFile(c echo.Context) error {
	request := &models.CreateFileDirRequest{}
	if err := c.Bind(request); err != nil {
		return resolveError(err, c, http.StatusBadRequest)
	}
	ctx := c.Request().Context()
	jwtClaims, err := l.Authen.GetDataViaToken(ctx, c.Get("user"))
	if err != nil {
		return resolveError(err, c, http.StatusInternalServerError)
	}
	newFileUUID, err := l.LocalFMan.CreateNewFile(ctx, jwtClaims.UserUUID, request.Name, request.ParentDirUUID)
	if err != nil {
		return resolveError(err, c, 0)
	}
	return c.JSON(http.StatusOK,
		models.SuccessFManResponse{
			EntityUUID: newFileUUID,
			Message:    "Created file successfully",
		},
	)
}

func (l *LocalFManHTTPHandler) getFile(c echo.Context) error {
	fileUUID := c.Param("id")
	ctx := c.Request().Context()
	jwtClaims, err := l.Authen.GetDataViaToken(ctx, c.Get("user"))
	if err != nil {
		return resolveError(err, c, http.StatusInternalServerError)
	}
	file, err := l.LocalFMan.GetFile(ctx, jwtClaims.UserUUID, fileUUID)
	if err != nil {
		return resolveError(err, c, 0)
	}
	return c.JSON(http.StatusOK, file)
}

func (l *LocalFManHTTPHandler) copyFile(c echo.Context) error {
	fileUUID := c.Param("id")
	req := &models.CopyMvRequest{}
	if err := c.Bind(req); err != nil {
		return resolveError(err, c, http.StatusBadRequest)
	}
	ctx := c.Request().Context()
	jwtClaims, err := l.Authen.GetDataViaToken(ctx, c.Get("user"))
	if err != nil {
		return resolveError(err, c, http.StatusInternalServerError)
	}
	newFileUUID, err := l.LocalFMan.CopyFile(ctx, jwtClaims.UserUUID, fileUUID, req.DstDirUUID)
	if err != nil {
		return resolveError(err, c, 0)
	}
	return c.JSON(http.StatusOK,
		models.SuccessFManResponse{
			EntityUUID: newFileUUID,
			Message:    "Copied file successfully",
		},
	)
}

func (l *LocalFManHTTPHandler) moveFile(c echo.Context) error {
	fileUUID := c.Param("id")
	req := &models.CopyMvRequest{}
	if err := c.Bind(req); err != nil {
		return resolveError(err, c, http.StatusBadRequest)
	}
	ctx := c.Request().Context()
	jwtClaims, err := l.Authen.GetDataViaToken(ctx, c.Get("user"))
	if err != nil {
		return resolveError(err, c, http.StatusInternalServerError)
	}
	err = l.LocalFMan.MoveFile(ctx, jwtClaims.UserUUID, fileUUID, req.DstDirUUID)
	if err != nil {
		return resolveError(err, c, 0)
	}
	return c.JSON(http.StatusOK,
		models.SuccessFManResponse{
			EntityUUID: fileUUID,
			Message:    "Moved file successfully",
		},
	)
}

func (l *LocalFManHTTPHandler) renameFile(c echo.Context) error {
	fileUUID := c.Param("id")
	req := &models.RenameRequest{}
	if err := c.Bind(req); err != nil {
		return resolveError(err, c, http.StatusBadRequest)
	}
	ctx := c.Request().Context()
	jwtClaims, err := l.Authen.GetDataViaToken(ctx, c.Get("user"))
	if err != nil {
		return resolveError(err, c, http.StatusInternalServerError)
	}
	err = l.LocalFMan.RenameFile(ctx, jwtClaims.UserUUID, fileUUID, req.NewName)
	if err != nil {
		return resolveError(err, c, 0)
	}
	return c.JSON(http.StatusOK,
		models.SuccessFManResponse{
			EntityUUID: fileUUID,
			Message:    "Renamed file successfully",
		},
	)
}

func (l *LocalFManHTTPHandler) downloadFile(c echo.Context) error {
	fileUUID := c.Param("id")
	ctx := c.Request().Context()
	jwtClaims, err := l.Authen.GetDataViaToken(ctx, c.Get("user"))
	if err != nil {
		return resolveError(err, c, http.StatusInternalServerError)
	}
	downloadPld, err := l.LocalFMan.DownloadFile(ctx, jwtClaims.UserUUID, fileUUID)
	if err != nil {
		return resolveError(err, c, 0)
	}
	defer downloadPld.File.Close()
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", "attachment", downloadPld.Filename))
	contentLength, err := downloadPld.File.GetSize()
	if err != nil {
		return resolveError(err, c, http.StatusInternalServerError)
	}
	c.Response().Header().Set(echo.HeaderContentLength, strconv.FormatInt(contentLength, 10))
	return c.Stream(http.StatusOK, echo.MIMEOctetStream, downloadPld.File)
}

func (l *LocalFManHTTPHandler) downloadFileBatch(c echo.Context) error {
	paramValues := c.Request().URL.Query()
	fileUUIDList := paramValues["fileUUID"]
	ctx := c.Request().Context()
	jwtClaims, err := l.Authen.GetDataViaToken(ctx, c.Get("user"))
	if err != nil {
		return resolveError(err, c, http.StatusInternalServerError)
	}
	downloadPld, err := l.LocalFMan.DownloadFileBatch(ctx, jwtClaims.UserUUID, fileUUIDList)
	if err != nil {
		return resolveError(err, c, 0)
	}
	defer func() {
		downloadPld.TmpFile.Close()
		downloadPld.TmpFile.Remove()
	}()

	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", "attachment", downloadPld.Filename))
	contentLength, err := downloadPld.TmpFile.GetSize()
	if err != nil {
		return resolveError(err, c, http.StatusInternalServerError)
	}
	c.Response().Header().Set(echo.HeaderContentLength, strconv.FormatInt(contentLength, 10))
	return c.Stream(http.StatusOK, echo.MIMEOctetStream, downloadPld.TmpFile)
}

func (l *LocalFManHTTPHandler) softRemoveFile(c echo.Context) error {
	fileUUID := c.Param("id")
	ctx := c.Request().Context()
	jwtClaims, err := l.Authen.GetDataViaToken(ctx, c.Get("user"))
	if err != nil {
		return resolveError(err, c, http.StatusInternalServerError)
	}
	err = l.LocalFMan.SoftRemoveFile(ctx, jwtClaims.UserUUID, fileUUID)
	if err != nil {
		return resolveError(err, c, 0)
	}
	return c.JSON(http.StatusOK,
		models.SuccessFManResponse{
			EntityUUID: fileUUID,
			Message:    "Removed file successfully",
		},
	)
}

func (l *LocalFManHTTPHandler) hardRemoveFile(c echo.Context) error {
	fileUUID := c.Param("id")
	ctx := c.Request().Context()
	jwtClaims, err := l.Authen.GetDataViaToken(ctx, c.Get("user"))
	if err != nil {
		return resolveError(err, c, http.StatusInternalServerError)
	}
	err = l.LocalFMan.HardRemoveFile(ctx, jwtClaims.UserUUID, fileUUID)
	if err != nil {
		return resolveError(err, c, 0)
	}
	return c.JSON(http.StatusOK,
		models.SuccessFManResponse{
			EntityUUID: fileUUID,
			Message:    "Removed file successfully",
		},
	)
}

func (l *LocalFManHTTPHandler) createDir(c echo.Context) error {
	request := &models.CreateFileDirRequest{}
	if err := c.Bind(request); err != nil {
		return resolveError(err, c, http.StatusBadRequest)
	}
	ctx := c.Request().Context()
	jwtClaims, err := l.Authen.GetDataViaToken(ctx, c.Get("user"))
	if err != nil {
		return resolveError(err, c, http.StatusInternalServerError)
	}
	newDirUUID, err := l.LocalFMan.CreateNewDirectory(ctx, jwtClaims.UserUUID, request.Name, request.ParentDirUUID)
	if err != nil {
		return resolveError(err, c, 0)
	}
	return c.JSON(http.StatusOK,
		models.SuccessFManResponse{
			EntityUUID: newDirUUID,
			Message:    "Created directory successfully",
		},
	)
}

func (l *LocalFManHTTPHandler) getDirectory(c echo.Context) error {
	dirUUID := c.Param("id")
	ctx := c.Request().Context()
	jwtClaims, err := l.Authen.GetDataViaToken(ctx, c.Get("user"))
	if err != nil {
		return resolveError(err, c, http.StatusInternalServerError)
	}
	if dirUUID != "" {
		dir, err := l.LocalFMan.GetDirectory(ctx, jwtClaims.UserUUID, dirUUID)
		if err != nil {
			return resolveError(err, c, 0)
		}
		return c.JSON(http.StatusOK, dir)
	}
	rootDir, err := l.LocalFMan.GetRootDirectory(ctx, jwtClaims.UserUUID)
	if err != nil {
		return resolveError(err, c, 0)
	}
	return c.JSON(http.StatusOK, rootDir)
}
