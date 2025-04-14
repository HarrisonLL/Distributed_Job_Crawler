package utils

import (
	"context"
	"fmt"
	"go_services/database"
	"go_services/models"
	"io"
	"log"
	"os"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// Debug mode allows preserve ended docker container
// Then use "docker logs <container-id>" to debug
func RunDockerContainer(envVars []string, volumeMappings []string, cmd []string, jobType models.JobType, taskId string, debug bool, releaseToken func(), wg *sync.WaitGroup, blocking bool) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Printf("Failed to start crawler for company %s: %v", jobType.CompanyName, err)
		return
	}

	hostConfig := &container.HostConfig{
		Binds: volumeMappings,
	}
	config := &container.Config{
		Image: jobType.DockerImageID,
		Cmd:   cmd,
		Env:   envVars,
	}
	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"jc_network": {},
		},
	}

	resp, err := cli.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, nil, "")
	if err != nil {
		log.Printf("Failed to start crawler for company %s: %v", jobType.CompanyName, err)
		return
	}
	if err := cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Printf("Failed to start crawler for company %s: %v", jobType.CompanyName, err)
		return
	}
	log.Printf("Container %s started successfully\n", resp.ID)
	// Create task record to DB
	log.Printf("Started container %s for company %s", resp.ID, jobType.CompanyName)
	err = database.CreateTask(taskId, resp.ID, jobType.CompanyName, jobType.JobTypeName, "USA")
	if err != nil {
		log.Printf("Failed to create task for company %s: %v", jobType.CompanyName, err)
	}

	if blocking {
		// Wait synchronously
		statusCh, errCh := cli.ContainerWait(context.Background(), resp.ID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				log.Printf("Error while waiting for container %s: %v", resp.ID, err)
				return
			}
		case <-statusCh:
			handleContainerFinish(cli, resp.ID)
		}
	} else if !debug && wg != nil {
		go func(containerID string) {
			defer wg.Done()
			defer releaseToken()
			statusCh, errCh := cli.ContainerWait(context.Background(), containerID, container.WaitConditionNotRunning)
			select {
			case err := <-errCh:
				if err != nil {
					log.Printf("Error while waiting for container %s: %v", containerID, err)
					return
				}
			case <-statusCh:
				handleContainerFinish(cli, containerID)
			}
		}(resp.ID)
	}
}

// Save logs only when container exit with error
func handleContainerFinish(cli *client.Client, containerID string) {
	ctx := context.Background()
	inspect, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		log.Printf("Error inspecting container %s: %v", containerID, err)
		return
	}

	if inspect.State.ExitCode != 0 {
		logs, logErr := cli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
		if logErr == nil {
			defer logs.Close()
			logData, _ := io.ReadAll(logs)
			logStoragePath := os.Getenv("WORKER_LOG_STORAGE_PATH")
			if logStoragePath != "" {
				logFilePath := fmt.Sprintf("%s/container_%s.log", logStoragePath, containerID)
				err := os.WriteFile(logFilePath, logData, 0644)
				if err != nil {
					log.Printf("Failed to write logs to %s: %v", logFilePath, err)
				} else {
					log.Printf("Saved error logs for container %s at %s", containerID, logFilePath)
				}
			}
			database.UpdateTaskStatus("", containerID, 3)
		} else {
			log.Printf("Failed to fetch logs for container %s: %v", containerID, logErr)
		}
	} else {
		log.Printf("Container %s finished successfully", containerID)
	}

	if err := cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{}); err != nil {
		log.Printf("Failed to remove container %s: %v", containerID, err)
	} else {
		log.Printf("Removed container %s", containerID)
	}
}

func PullDockerImage(imageName string) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}

	// Check if the image already exists
	imageList, err := cli.ImageList(context.Background(), types.ImageListOptions{
		Filters: filters.NewArgs(filters.Arg("reference", imageName)),
	})
	if err != nil {
		return "", err
	}

	if len(imageList) > 0 {
		log.Printf("Image %s already exists. Skipping pull.", imageName)
		return imageList[0].ID, nil
	}

	// else pull image
	out, err := cli.ImagePull(context.Background(), imageName, types.ImagePullOptions{})
	if err != nil {
		return "", err
	}
	defer out.Close()

	imageInspect, _, err := cli.ImageInspectWithRaw(context.Background(), imageName)
	if err != nil {
		return "", err
	}

	return imageInspect.ID, nil
}
