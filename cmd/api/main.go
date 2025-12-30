package main

import (
	appfx "Fynance/internal/fx"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		appfx.AppModule,
	).Run()
}
