package utils

import (
	"context"
	"fmt"
	"go_services/database"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// Debug allows preserve end docker container
// So to use "docker logs <container-id>" to debug
func RunDockerContainer(image string, envVars []string, volumeMappings []string, cmd []string, debug bool) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}

	hostConfig := &container.HostConfig{
		Binds: volumeMappings,
	}
	config := &container.Config{
		Image: image,
		Cmd:   cmd,
		Env:   envVars,
	}

	resp, err := cli.ContainerCreate(context.Background(), config, hostConfig, nil, nil, "")
	if err != nil {
		return "", err
	}
	if err := cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}
	log.Printf("Container %s started successfully\n", resp.ID)

	if !debug {
		// Start a goroutine to wait for the container to finish and then remove it,
		// ContainerWait API will notify current goroutine thread by starting two channels (statusCh, errCh)
		// Save logs only when container exit with error
		go func(containerID string) {
			statusCh, errCh := cli.ContainerWait(context.Background(), containerID, container.WaitConditionNotRunning)
			select {
			case err := <-errCh:
				if err != nil {
					log.Printf("Error while waiting for container %s: %v", containerID, err)
					return
				}
			case <-statusCh:
				ctx := context.Background()
				inspect, err := cli.ContainerInspect(ctx, containerID)
				if err != nil {
					log.Printf("Error while inspecting container %s: %v", containerID, err)
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

				// Remove exited container
				if err := cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{}); err != nil {
					log.Printf("Failed to remove container %s: %v", containerID, err)
				} else {
					log.Printf("Removed container %s", containerID)
				}
			}
		}(resp.ID)
	}
	return resp.ID, nil
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
