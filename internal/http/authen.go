package http

import (
	"github.com/labstack/echo/v4"
	"github.com/nvthongswansea/xtreme/internal/authen"
	"github.com/nvthongswansea/xtreme/internal/models"
	"net/http"
)

type authenHTTPHandler struct {
	authen authen.Authenticator
}

func attachAuthenHTTPHandlerHandler(e *echo.Echo, a authen.Authenticator) {
	handler := &authenHTTPHandler{authen: a}
	e.POST("/register", handler.registerUser)
	e.POST("/login", handler.loginUser)
}

func (a *authenHTTPHandler) registerUser(c echo.Context) error {
	request := &models.CreateUserRequest{}
	if err := c.Bind(request); err != nil {
		return resolveError(err, c, http.StatusBadRequest)
	}
	ctx := c.Request().Context()
	err := a.authen.Register(ctx, request.Username, request.Password)
	if err != nil {
		return resolveError(err, c, 0)
	}
	return c.JSON(http.StatusOK,
		models.SuccessRegisterResponse{
			Message: "Registered user successful",
		},
	)
}

func (a *authenHTTPHandler) loginUser(c echo.Context) error {
	request := &models.AuthenUserRequest{}
	if err := c.Bind(request); err != nil {
		return resolveError(err, c, http.StatusBadRequest)
	}
	ctx := c.Request().Context()
	token, err := a.authen.Login(ctx, request.Username, request.Password)
	if err != nil {
		return resolveError(err, c, 0)
	}
	return c.JSON(http.StatusOK,
		models.SuccessAuthenResponse{
			Token:   token,
			Message: "Logged in successful",
		},
	)
}
