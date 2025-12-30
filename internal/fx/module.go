package fx

import "go.uber.org/fx"

// AppModule reúne todos os módulos da aplicação
var AppModule = fx.Options(
	ConfigModule,
	InfrastructureModule,
	DomainModule,
	MiddlewareModule,
	RoutesModule,
	ServerModule,
)
