package main

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/client"
)

func runHealthcheck() int {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("docker client error: %v\n", err)
		return 1
	}
	defer dockerClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = dockerClient.Ping(ctx)
	if err != nil {
		fmt.Printf("docker ping failed: %v\n", err)
		return 1
	}

	fmt.Println("ok")
	return 0
}
