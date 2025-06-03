package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

func RunCLI() {
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage:")
		fmt.Println("go run main.go -service CLI <command> [arguments]")
		fmt.Println("Commands:")
		fmt.Println("  delete <company_name> <job_type_name>")
		fmt.Println("  add <company_name> <job_type_name> <docker_image_name>")
		fmt.Println("  setup_ttl")
		os.Exit(1)
	}

	command := args[0]
	switch command {
	case "delete_job_type":
		if len(args) != 3 {
			fmt.Println("Delete usage: go run main.go -service CLI delete <company_name> <job_type_name>")
			os.Exit(1)
		}
		job := JobType{
			CompanyName: args[1],
			JobTypeName: args[2],
		}
		deleteJobType(job)
	case "add_job_type":
		if len(args) != 4 {
			fmt.Println("Add usage: go run main.go -service CLI add <company_name> <job_type_name> <docker_image_name>")
			os.Exit(1)
		}
		job := JobType{
			CompanyName:     args[1],
			JobTypeName:     args[2],
			DockerImageName: args[3],
		}
		insertUpdateJobType(job)
	case "add_update_ttl":
		if len(args) != 2 {
			fmt.Println("Add usage: go run main.go -service CLI add_update_ttl <days>")
			os.Exit(1)
		}
		days, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("Invalid number of days.")
			os.Exit(1)
		}
		if err := addUpdateTTLForAllJobcrawlerDatabases(days); err != nil {
			fmt.Printf("Failed to set TTL indexes: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("TTL setup completed successfully.")
	case "list_ttl":
		if err := listTTLForAllJobcrawlerDatabases(); err != nil {
			fmt.Printf("Failed to list TTL indexes: %v\n", err)
			os.Exit(1)
		}
	case "delete_ttl":
		if err := deleteTTLForAllJobcrawlerDatabases(); err != nil {
			fmt.Printf("Failed to list TTL indexes: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: delete, add, setup_ttl")
		os.Exit(1)
	}
}
