package main

import (
	"flag"
	"go_services/cli"
	"go_services/database"
	"go_services/services"
	"log"
)

func main() {
	serviceMode := flag.String("service", "", "Service mode: taskSvc, userSvc, scheduler, emailConsumer, CLI")
	flag.Parse()

	switch *serviceMode {
	case "taskSvc":
		database.Init()
		services.StartTaskSvc()
	case "userSvc":
		database.InitUserSvc()
		services.StartUserSvc()
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
