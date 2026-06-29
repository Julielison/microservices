package main

import (
	"log"

	"github.com/Julielison/microservices/shipping/config"
	"github.com/Julielison/microservices/shipping/internal/adapters/db"
	grpcadapter "github.com/Julielison/microservices/shipping/internal/adapters/grpc"
	"github.com/Julielison/microservices/shipping/internal/application/core/api"
)

func main() {
	dbAdapter, err := db.NewAdapter(config.GetDataSourceURL())
	if err != nil {
		log.Fatalf("Failed to connect to database. Error: %v", err)
	}

	application := api.NewApplication(dbAdapter)
	grpcAdapter := grpcadapter.NewAdapter(application, config.GetApplicationPort())
	grpcAdapter.Run()
}
