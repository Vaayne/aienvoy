package middlerware

import (
	"context"
	"log/slog"
	"strings"

	"github.com/Vaayne/aienvoy/internal/core/auth"
	"github.com/Vaayne/aienvoy/internal/pkg/config"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/daos"
)

// AuthByApiKeyMiddleware is a middleware to auth user by api key
func AuthByApiKeyMiddleware(d *daos.Dao) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			val := c.Get(config.ContextKeyAuthRecord)
			if val == nil {
				authHeader := c.Request().Header.Get("Authorization")
				if authHeader != "" {
					auths := strings.Split(authHeader, " ")
					if len(auths) == 2 && auths[0] == "Bearer" {
						apiKeyStr := auths[1]

						authRecord, err := auth.FindAuthRecordByApiKey(context.TODO(), d, apiKeyStr)
						if err != nil {
							slog.Info("error get user by api key", "err", err, "key", authHeader)
							return apis.NewUnauthorizedError("invalid api key", nil)
						}
						c.Set(config.ContextKeyAuthRecord, authRecord)
						c.Set(config.ContextKeyUserId, authRecord.GetString("user_id"))
						c.Set(config.ContextKeyApiKey, apiKeyStr)
					}
				}
			}
			return next(c)
		}
	}
}
