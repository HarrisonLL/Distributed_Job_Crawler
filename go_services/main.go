package main

import (
	"flag"
	"go_services/cli"
	"go_services/database"
	"go_services/services"
	"log"
)

func main() {
	serviceMode := flag.String("service", "", "Service mode: web, scheduler, emailConsumer, CLI")
	flag.Parse()

	switch *serviceMode {
	case "web":
		database.Init()
		services.StartWeb()
	case "scheduler":
		database.Init()
		services.StartScheduler()
	case "emailConsumer":
		services.StartEmailConsumer()
	case "CLI":
		database.Init()
		cli.RunCLI()
	default:
		log.Fatalf("Invalid service mode: %s", *serviceMode)
	}

}
