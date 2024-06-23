package main

import (
	"social_server/src/app/api"
	. "social_server/src/utils/log"
)

func main() {
	SetupLogger()

	modAPi := api.NewModApi()
	err := modAPi.StartRpcServer()
	if err != nil {
		Log.Info("could not login: %v", err)
		return
	}
	Log.Info("Server Started.")
}
