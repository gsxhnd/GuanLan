package main

import "go.uber.org/fx"

func main() {
	fx.New(
		fx.Provide(
			NewConfig,
			NewListener,
			NewGRPCServer,
		),
		fx.Invoke(Run),
	).Run()
}
