package main

import "go.uber.org/fx"

func main() {
	fx.New(
		fx.Provide(
			NewConfig,
			NewStore,
			NewServices,
			NewScheduler,
			NewListener,
			NewGRPCServer,
		),
		fx.Invoke(Run),
	).Run()
}
