package main

import (
	"log"

	"github.com/juan131/api-mock/internal/service"
)

func main() {
	// Init service
	svc := service.NewService()
	if err := svc.LoadConfig(); err != nil {
		log.Fatal(err)
	}

	// Start service
	svc.MakeRouter()
	svc.ListenAndServe()
}
