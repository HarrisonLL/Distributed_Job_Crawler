package utils

import (
	"context"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func RunDockerContainer(image string, envVars []string, volumeMappings []string, cmd []string) (string, error) {
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
