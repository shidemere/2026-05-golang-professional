package internalhttp

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func hello(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{"message": "Hello from server!"})
}
