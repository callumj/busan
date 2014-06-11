package app

import (
	"github.com/callumj/docker-mate/remote"
	"github.com/callumj/docker-mate/utils"
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"strings"
)

type ContainerConfig struct {
	Attributes *docker.Config
	Image      VersionImage
}

type ActiveContainers struct {
	CurrentVersionId string
	OtherVersionIds  []string
}

func ParseContainerConfig(attributesFilePath string) (ContainerConfig, error) {
	fileContents, err := ioutil.ReadFile(attributesFilePath)
	if err != nil {
		return ContainerConfig{}, err
	}

	target := ContainerConfig{}
	attrTarget := docker.Config{}
	err = yaml.Unmarshal([]byte(fileContents), &attrTarget)
	if err != nil {
		return ContainerConfig{}, err
	}
	target.Attributes = &attrTarget

	return target, nil
}

func SpinUpContainer(conf ContainerConfig) error {
	res, err := GetActiveContainers(conf.Image)
	if err != nil {
		return err
	}

	if len(res.CurrentVersionId) != 0 {
		utils.LogMessage("Latest container is already running %s\r\n", res.CurrentVersionId)
	} else {
		if len(res.OtherVersionIds) > 0 {
			// drop the other containers
			RemoveContainers(res.OtherVersionIds)
		}

		// deploy the container
		CreateContainer(conf)
	}

	return nil
}

// Provides ability to see if we are running the latest container
func GetActiveContainers(img VersionImage) (ActiveContainers, error) {
	allContainers, err := remote.DockerClient.ListContainers(docker.ListContainersOptions{All: true})

	if err != nil {
		return ActiveContainers{}, err
	}

	structr := ActiveContainers{}

	for _, con := range allContainers {
		if con.Image == utils.BuildName(img.Version) {
			structr.CurrentVersionId = con.ID
		} else if strings.Contains(con.Image, utils.GlobalOptions.Name) {
			structr.OtherVersionIds = append(structr.OtherVersionIds, con.ID)
		}
	}

	return structr, nil
}

func RemoveContainers(containerIds []string) error {
	for _, conId := range containerIds {
		err := remote.DockerClient.StopContainer(conId, 90)
		if err != nil {
			return err
		}

		_, err = remote.DockerClient.WaitContainer(conId)
		if err != nil {
			return nil
		}

		err = remote.DockerClient.RemoveContainer(docker.RemoveContainerOptions{ID: conId})
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateContainer(conf ContainerConfig) {
	utils.LogMessage("%v\n\n", conf)
}
