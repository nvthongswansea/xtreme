package http

//import (
//	"net/http"
//
//	"github.com/labstack/echo/v4"
//	"github.com/nvthongswansea/xtreme/internal/fman"
//)
//
//// ResponseError represents http response error in JSON format
//type Response struct {
//	Message string `json:"message"`
//}
//
//// FmanHandler represents the http handler for file manage
//type FmanHandler struct {
//	FmanUsecase fman.FmanUsecase
//}
//
//// InitFmanHandler initialize file manager endpoints
//func InitFmanHandler(e *echo.Echo, uc fman.FmanUsecase) {
//	handler := &FmanHandler{FmanUsecase: uc}
//	g := e.Group("/fman")
//	g.POST("/file", handler.UploadFile)
//}
//
//func (h *FmanHandler) UploadFile(c echo.Context) error {
//	// Get file from form
//	file, err := c.FormFile("file")
//	if err != nil {
//		return err
//	}
//	src, err := file.Open()
//	if err != nil {
//		return err
//	}
//	defer src.Close()
//
//	filename := c.FormValue("filename")
//	parentUUID := c.FormValue("parent_uuid")
//	// Save file
//	err = h.FmanUsecase.UploadFile(filename, parentUUID, src)
//	if err != nil {
//		return err
//	}
//	return c.JSON(http.StatusOK, Response{Message: "Uploaded file successfully"})
//}
