package device

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

// Get all the connected devices to the host by reading the symlinks in /dev
func getConnectedDevices() ([]string, error) {
	// Get all files/symlinks/folders in /dev
	var connectedDevices []string = []string{}
	devFiles, err := filepath.Glob("/dev/*")
	if err != nil {
		fmt.Println("Error listing files in /dev:", err)
		return nil, err
	}

	for _, devFile := range devFiles {
		// Split the devFile to get only the file name
		_, fileName := filepath.Split(devFile)
		// If the filename is a device symlink
		// Add it to the connected devices list
		if strings.Contains(fileName, "device") {
			connectedDevices = append(connectedDevices, fileName)
		}
	}

	return connectedDevices, nil
}

var cli *client.Client

// Create a docker client singleton to be used by the provider
// This avoids exhausting docker socket connections and also makes code cleaner
// Might be changed in the future if this becomes a problem
func initDockerClient() error {
	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "get_host_containers",
		}).Error(". Error: " + err.Error())
		return err
	}

	return nil
}

// Get list of all containers on host
func getHostContainers() ([]types.Container, error) {
	if cli == nil {
		err := initDockerClient()
		if err != nil {
			return []types.Container{}, err
		}
	}

	// Get the list of containers
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		log.WithFields(log.Fields{
			"event": "get_host_containers",
		}).Error(". Error: " + err.Error())
		return nil, errors.New("Could not get container list: " + err.Error())
	}
	return containers, nil
}

// Check if device has an existing container
func (device *Device) hasContainer(allContainers []types.Container) (bool, error) {
	for _, container := range allContainers {
		// Parse plain container name
		containerName := strings.Replace(container.Names[0], "/", "", -1)

		if strings.Contains(containerName, device.UDID) {
			deviceContainer := DeviceContainer{
				ContainerID:     container.ID,
				ContainerStatus: container.Status,
				ImageName:       container.Image,
				ContainerName:   containerName,
			}
			device.Container = &deviceContainer
			return true, nil
		}
	}
	return false, nil
}

// Get a device pointer from Config for a device by udid
func GetDeviceByUDID(udid string) *Device {
	for _, device := range Config.Devices {
		if device.UDID == udid {
			return device
		}
	}

	return nil
}
