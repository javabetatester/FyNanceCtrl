package fx

import (
	"Fynance/config"
	"Fynance/internal/domain/user"
	"Fynance/internal/middleware"

	"go.uber.org/fx"
)

var MiddlewareModule = fx.Module("middleware",
	fx.Provide(
		newJwtService,
	),
)

func newJwtService(cfg *config.Config, userSvc *user.Service) (*middleware.JwtService, error) {
	return middleware.NewJwtService(cfg.JWT, userSvc)
}
