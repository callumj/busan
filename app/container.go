package app

import (
	"fmt"
	"github.com/callumj/busan/remote"
	"github.com/callumj/busan/utils"
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"strings"
)

type ContainerCreateConfig struct {
	Volumes      map[string]string
	ExposedPorts []string
}

type ContainerConfig struct {
	Attributes *ContainerCreateConfig
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
	attrTarget := ContainerCreateConfig{}
	err = yaml.Unmarshal([]byte(fileContents), &attrTarget)
	if err != nil {
		return ContainerConfig{}, err
	}
	target.Attributes = &attrTarget

	return target, nil
}

func SpinUpContainer(conf ContainerConfig) error {
	res, err := GetInstalledContainers(conf.Image)
	if err != nil {
		return err
	}

	if len(res.CurrentVersionId) != 0 {
		utils.LogMessage("Latest container is already installed %s\r\n", res.CurrentVersionId)
	} else {
		if len(res.OtherVersionIds) > 0 {
			// drop the other containers
			err = RemoveContainers(res.OtherVersionIds)
			if err != nil {
				return err
			}
		}

		// deploy the container
		utils.LogMessage("Creating container\r\n")
		resp, err := CreateContainer(conf)

		if err != nil {
			return err
		}

		res = resp
	}

	// boot up the container if needed
	isRunning, err := IsContainerRunning(res.CurrentVersionId)
	if err != nil {
		return err
	}

	if isRunning {
		utils.LogMessage("Container %s is already running\r\n", res.CurrentVersionId)
	} else {
		err = StartContainer(res.CurrentVersionId, conf.Attributes)
		if err == nil {
			utils.LogMessage("Container %s is now up\r\n", res.CurrentVersionId)
		}
	}

	return err
}

// Provides ability to see if we are running the latest container
func GetInstalledContainers(img VersionImage) (ActiveContainers, error) {
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

func CreateContainer(conf ContainerConfig) (ActiveContainers, error) {
	nativeConf := docker.Config{}
	if conf.Attributes != nil {
		volumes := make(map[string]struct{})
		var empty struct{}
		for vol, _ := range conf.Attributes.Volumes {
			volumes[vol] = empty
		}
		nativeConf.Volumes = volumes
	}

	nativeConf.Image = conf.Image.Image.ID
	nativeConf.Env = []string{fmt.Sprintf("NAME=%s", utils.GlobalOptions.Name), fmt.Sprintf("VERSION=%s", conf.Image.Version)}
	opts := docker.CreateContainerOptions{
		Name:   utils.GlobalOptions.Name,
		Config: &nativeConf,
	}
	resp, err := remote.DockerClient.CreateContainer(opts)

	if err != nil {
		return ActiveContainers{}, err
	}

	return ActiveContainers{CurrentVersionId: resp.ID}, err
}

func IsContainerRunning(conId string) (bool, error) {
	running, err := remote.DockerClient.ListContainers(docker.ListContainersOptions{All: false})

	if err != nil {
		return false, err
	}

	for _, con := range running {
		if con.ID == conId {
			return true, nil
		}
	}

	return false, nil
}

func StartContainer(conId string, config *ContainerCreateConfig) error {
	nativeConf := docker.HostConfig{}
	ports := make(map[docker.Port][]docker.PortBinding)
	if config != nil {
		var opts []string
		for target, source := range config.Volumes {
			splitted := strings.Split(source, ":")
			var built string
			if len(splitted) == 1 {
				built = fmt.Sprintf("%v:%v", splitted[0], target)
			} else {
				built = fmt.Sprintf("%v:%v:%v", splitted[0], target, splitted[1])
			}
			opts = append(opts, built)
		}
		nativeConf.Binds = opts

		for _, p := range config.ExposedPorts {
			castedPort := docker.Port(p)
			ports[castedPort] = []docker.PortBinding{}
		}
	}
	nativeConf.PortBindings = ports

	err := remote.DockerClient.StartContainer(conId, &nativeConf)
	return err
}
