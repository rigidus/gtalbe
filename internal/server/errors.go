package server

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

var ErrInvalidAddress = echo.NewHTTPError(http.StatusBadRequest, "invalid address")
