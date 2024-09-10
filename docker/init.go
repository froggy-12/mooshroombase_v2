package docker

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/froggy-12/mooshroombase_v2/configs"
	"github.com/froggy-12/mooshroombase_v2/utils"
)

func Init() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		log.Fatal("Error Creating API client for docker package: ", err.Error())
	}

	requriedImages := []string{"redis:latest", "mongo:latest", "mariadb:latest"}
	utils.DebugLogger("docker", "checking installed Images")
	installedImages, err := cli.ImageList(context.Background(), image.ListOptions{All: true})

	if err != nil {
		log.Fatal("Error Listing all installed Images: ", err.Error())
	}

	missingImages := []string{}
	utils.DebugLogger("docker", "finding missing images 📲")
	for _, requiredImage := range requriedImages {
		found := false
		for _, installedImage := range installedImages {
			if requiredImage == installedImage.RepoTags[0] {
				found = true
				break
			}
		}
		if !found {
			missingImages = append(missingImages, requiredImage)
		}
	}

	if len(missingImages) > 0 {
		utils.DebugLogger("docker", "found some missing Images")
		for _, missingImage := range missingImages {
			utils.DebugLogger("docker", "Found Missing Image: "+missingImage)
			reader, err := cli.ImagePull(context.Background(), missingImage, image.PullOptions{})
			if err != nil {
				log.Fatal("Error Pulling Image: ", missingImage)
			} else {
				io.Copy(io.Discard, reader)
				utils.DebugLogger("docker", "Successfully pulled: "+missingImage)
			}
		}
	} else {
		utils.DebugLogger("docker", "all required images are present 👍👍👍")
	}

	requiredContainers := []string{}

	utils.DebugLogger("docker", "looking for required containers 📦")
	for _, container := range configs.Configs.DatabaseConfigurations.RunningDatabases {
		switch container {
		case "mongodb":
			requiredContainers = append(requiredContainers, "mooshroombase-mongo")
		case "redis":
			requiredContainers = append(requiredContainers, "mooshroombase-redis")
		case "mariadb":
			requiredContainers = append(requiredContainers, "mooshroombase-mariadb")
		}
	}

	utils.DebugLogger("docker", "looking for installed containers 📦")
	installedContainers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})

	if err != nil {
		log.Fatal("Error getting all containers on system: ", err.Error())
	}

	missingContainers := []string{}

	utils.DebugLogger("docker", "finding missing containers 📦")
	for _, requiredContainer := range requiredContainers {
		found := false
		for _, hostedContainer := range installedContainers {
			if strings.Split(hostedContainer.Names[0], "/")[1] == requiredContainer {
				found = true
				break
			}
		}
		if !found {
			missingContainers = append(missingContainers, requiredContainer)
			utils.DebugLogger("docker", "missing container found 📦: "+requiredContainer)
		}

	}

	if len(missingContainers) > 0 {
		utils.DebugLogger("docker", "creating missing container 📦")
		for _, container := range missingContainers {
			switch container {
			case "mooshroombase-mongo":
				err = createAndStartMongoDBContainer(cli, "mooshroombase-mongo", configs.Configs.DatabaseConfigurations.MongoDBServerPort, "mongo:latest", configs.Configs.DatabaseConfigurations.MongoDBRootPassword)
			case "mooshroombase-redis":
				err = createAndStartRedisDBContainer(cli, "mooshroombase-redis", configs.Configs.DatabaseConfigurations.RedisDBServerPort, "redis:latest", configs.Configs.DatabaseConfigurations.RedisDBRootPassword)
			case "mooshroombase-mariadb":
				err = createAndStartMariaDBContainer(cli, "mooshroombase-mariadb", configs.Configs.DatabaseConfigurations.MariaDBServerPort, "mariadb:latest", configs.Configs.DatabaseConfigurations.MariaDBRootPassword)
			}
			if err != nil {
				log.Fatal("Error creating and starting container: ", err.Error())
			}
		}
	} else {
		utils.DebugLogger("docker", "no missing container found to create")
	}

	utils.DebugLogger("docker", "checking if any container sleeping")
	for _, requiredContainer := range requiredContainers {
		for _, installedContainer := range installedContainers {
			if strings.Split(installedContainer.Names[0], "/")[1] == requiredContainer {
				if installedContainer.State == "exited" {
					utils.DebugLogger("docker", "found container: "+requiredContainer+" is sleeping 😪 starting again")
					err = cli.ContainerStart(context.Background(), installedContainer.ID, container.StartOptions{})
					if err != nil {
						log.Fatal("Error Starting Container: ", requiredContainer+" error: "+err.Error())
					}
					break
				}
			}
		}
	}

	utils.DebugLogger("docker", "all docker containers running perfectly 🐋👍")

	// stopping thread for 2 mins because sometime database containers takes time to handle connections
	utils.DebugLogger("docker", "stopping main thread for configuring containers")

	ticker := time.NewTicker(time.Second)
	done := make(chan bool)
	go func() {
		start := time.Now()
		for range ticker.C {
			timeLeft := int(float64(configs.Configs.ExtraConfigurations.AfterStartDockerThreadRestartTime) - time.Since(start).Seconds())
			if timeLeft <= 0 {
				break
			}
			utils.DebugLogger("docker", fmt.Sprintf("Waiting for containers to be cold... %d seconds left", timeLeft))
		}
		done <- true
	}()
	<-done
	ticker.Stop()

	utils.DebugLogger("docker", "thread started again 😊")

}

func createAndStartMongoDBContainer(cli *client.Client, name string, port string, image string, password string) error {
	utils.DebugLogger("docker", "Creating and Starting MongoDB server")
	containerConfig := &container.Config{
		Image: image,
		ExposedPorts: nat.PortSet{
			nat.Port("27017"): struct{}{},
		},
		Env: []string{
			"MONGO_INITDB_ROOT_USERNAME=" + "root",
			"MONGO_INITDB_ROOT_PASSWORD=" + password,
		},
	}
	hostConfig := container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{
			nat.Port("27017"): {
				{
					HostIP:   "0.0.0.0",
					HostPort: port,
				},
			},
		},
	}
	containerName := name
	cont, err := cli.ContainerCreate(context.Background(), containerConfig, &hostConfig, nil, nil, containerName)
	if err != nil {
		return err
	}
	err = cli.ContainerStart(context.Background(), cont.ID, container.StartOptions{})
	return err
}

func createAndStartRedisDBContainer(cli *client.Client, name string, port string, image string, redisPassword string) error {
	utils.DebugLogger("docker", "Creating and Starting Redis server")

	containerConfig := &container.Config{
		Image: image,
		ExposedPorts: nat.PortSet{
			nat.Port("6379"): struct{}{},
		},
		Env: []string{
			"REDIS_PASSWORD=" + redisPassword,
		},
	}
	hostConfig := container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{
			nat.Port("6379"): {
				{
					HostIP:   "0.0.0.0",
					HostPort: port,
				},
			},
		},
	}
	containerName := name
	cont, err := cli.ContainerCreate(context.Background(), containerConfig, &hostConfig, nil, nil, containerName)
	if err != nil {
		return err
	}
	err = cli.ContainerStart(context.Background(), cont.ID, container.StartOptions{})
	return err
}

func createAndStartMariaDBContainer(cli *client.Client, name string, port string, image string, password string) error {
	utils.DebugLogger("docker", "Creating and Starting MariaDB server")

	containerConfig := &container.Config{
		Image: image,
		ExposedPorts: nat.PortSet{
			nat.Port("3306"): struct{}{},
		},
		Env: []string{
			"MARIADB_ROOT_PASSWORD=" + password,
		},
	}
	hostConfig := container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{
			nat.Port("3306"): {
				{
					HostIP:   "0.0.0.0",
					HostPort: port,
				},
			},
		},
	}
	containerName := name
	cont, err := cli.ContainerCreate(context.Background(), containerConfig, &hostConfig, nil, nil, containerName)
	if err != nil {
		return err
	}
	err = cli.ContainerStart(context.Background(), cont.ID, container.StartOptions{})
	ticker := time.NewTicker(time.Second)
	done := make(chan bool)
	go func() {
		start := time.Now()
		for range ticker.C {
			timeLeft := int(50 - time.Since(start).Seconds())
			if timeLeft <= 0 {
				break
			}
			utils.DebugLogger("docker", fmt.Sprintf("Waiting till mariadb container configured... %d seconds left", timeLeft))
		}
		done <- true
	}()
	<-done
	ticker.Stop()
	return err
}
