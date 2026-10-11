package internalhttp

import (
	"time"

	"github.com/labstack/echo/v4"
)

func loggingMiddleware(logger Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			start := time.Now()

			req := ctx.Request()

			err := next(ctx)

			res := ctx.Response()
			status := res.Status
			if status == 0 {
				status = 200
			}

			logger.Info(
				"request handled",
				"ip", ctx.RealIP(),
				"time", start,
				"method", req.Method,
				"path", req.URL.RequestURI(),
				"proto", req.Proto,
				"status", status,
				"latency", time.Since(start),
				"user_agent", req.UserAgent(),
			)

			return err
		}
	}
}
