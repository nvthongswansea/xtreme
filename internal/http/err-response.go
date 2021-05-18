package http

import (
	"github.com/labstack/echo/v4"
	"github.com/nvthongswansea/xtreme/internal/models"
	"net/http"
)

// resolveError resolves error.
func resolveError(err error, c echo.Context, forceCode int) error {
	if forceCode != 0 {
		return c.JSON(forceCode, nil)
	}
	switch err.(models.XtremeError).Code {
	case models.InternalServerErrorCode:
		return c.JSON(http.StatusInternalServerError, err)
	case models.BadInputErrorCode:
		return c.JSON(http.StatusBadRequest, err)
	case models.NotFoundErrorCode:
		return c.JSON(http.StatusNotFound, err)
	case models.ForbiddenOperationErrorCode:
		return c.JSON(http.StatusUnauthorized, err)
	default:
	}
	return c.JSON(http.StatusBadRequest, err)
}
