package core

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
)

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

func CheckDockerFile(dir string) error {
	dockerFilePath := fmt.Sprintf("%s/Dockerfile", dir)
	info, err := os.Stat(dockerFilePath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("dockerfile is a directory")
	}
	if !info.Mode().IsRegular() {
		return errors.New("dockerfile is not a regular file")
	}
	return nil
}

func BuildAndRunDockerContainer(dir string, appId string) error {
	appId = strings.ToLower(appId)
	if err := CheckDockerFile(dir); err != nil {
		return err
	}

	tar, err := archive.TarWithOptions(dir, &archive.TarOptions{})
	if err != nil {
		return err
	}

	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{appId},
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	res, err := cli.ImageBuild(context.Background(), tar, opts)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	err = print(res.Body, appId)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	cInfo, err := cli.ContainerInspect(context.Background(), appId)
	if err == nil {
		cli.ContainerRemove(context.Background(), cInfo.ID, types.ContainerRemoveOptions{
			Force: true,
		})
	}

	exposedPorts, portBindings, _ := nat.ParsePortSpecs([]string{
		"127.0.0.1:3002:3002",
	})
	cli.ContainerCreate(context.Background(), &container.Config{
		Image:        fmt.Sprintf("%s:latest", appId),
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		PortBindings: portBindings, // it supposed to be nat.PortMap
	}, nil, nil, appId)

	err = cli.ContainerStart(context.Background(), appId, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	CleanDocker()
	return nil
}

func print(rd io.Reader, appId string) error {

	token := make([]byte, 4)
	rand.Read(token)
	timestamp := time.Now().Unix()
	os.Mkdir("logs", 0777)
	// open output file
	fo, err := os.Create(fmt.Sprintf("logs/%s_docker_%s_%s.log", appId, fmt.Sprintf("%v", timestamp), fmt.Sprintf("%x", token)))
	if err != nil {
		return err
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// make a buffer to keep chunks that are read
	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := rd.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := fo.Write(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}

func CleanDocker() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: filters.NewArgs(
			filters.Arg("status", "exited"),
			filters.Arg("status", "created"),
		),
	})

	if err != nil {
		fmt.Println(err.Error())
	}

	for _, container := range containers {
		if strings.Contains(container.Image, "sha256") {
			fmt.Printf("Removing container %s\n", container.Names[0])
			cli.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{
				Force: true,
			})
			fmt.Printf("Removing related Image %s\n", container.Image)
			cli.ImageRemove(context.Background(), container.ImageID, types.ImageRemoveOptions{
				Force: true,
			})
		}
	}

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		fmt.Println(err.Error())
	}

	for _, image := range images {
		if image.RepoTags[0] == "<none>:<none>" {
			fmt.Printf("Removing image %s\n", image.ID)
			cli.ImageRemove(context.Background(), image.ID, types.ImageRemoveOptions{
				Force:         true,
				PruneChildren: true,
			})
		}
	}

	return nil
}
