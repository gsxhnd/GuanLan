package main

import "go.uber.org/fx"

func main() {
	fx.New(
		fx.Provide(
			NewConfig,
			NewStore,
			NewBiz,
			NewPythonClients,
			NewPredictExecutor,
			NewServices,
			NewScheduler,
			NewCron,
			NewHTTPServer,
		),
		fx.Invoke(Run),
	).Run()
}
