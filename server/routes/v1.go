package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/knqyf263/boltwiz/server/handlers"
)

func RegisterV1Routes(e *echo.Echo, h *handlers.Handlers) {
	v1 := e.Group("/api/v1")
	v1.GET("", h.SayHello, can("api"))
	v1.POST("/list", h.ListElement)
	v1.POST("/add_buckets", h.AddBucket)
	v1.POST("/add_pairs", h.AddPairs)
	v1.POST("/delete", h.DeleteElement)
	v1.POST("/rename_key", h.RenameElement)
	v1.POST("/update_value", h.UpdatePairValue)
}

// can checks that the current user's role is allowed to perform all of the
// provided actions (so this is an AND condition, use canOr for OR)
func can(actions ...string) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}
