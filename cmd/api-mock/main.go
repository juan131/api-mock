package main

import (
	"log"

	"github.com/juan131/api-mock/internal/service"
)

func main() {
	// Init service
	svcCfg, err := service.LoadConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	svc := service.Make(svcCfg)
	svc.LogConfiguration()
	svc.MakeRouter()

	// Start service
	svc.ListenAndServe()
}
