package app

import (
	"fmt"
	"github.com/callumj/docker-mate/remote"
	"github.com/callumj/docker-mate/utils"
	"github.com/jessevdk/go-flags"
)

func Run(args []string) {
	args, err := flags.ParseArgs(&utils.GlobalOptions, args)

	if err != nil {
		utils.QuitFatal()
	}

	if len(args) == 1 {
		utils.LogMessage("Usage: %s [OPTIONS] DOCKER_FILE_DIRECTORY", args[0])
		utils.QuitFatal()
	}

	remote.ConfigureDockerEndpoint(utils.GlobalOptions.HostAddress)

	dockerFileDirectory := args[1]
	err = beginMigrating(dockerFileDirectory)
	if err != nil {
		utils.LogMessage("Error: %v", err)
		utils.QuitFatal()
	} else {
		utils.QuitSuccess()
	}
}

func beginMigrating(dockerFileDirectory string) error {
	dockerFile, err := CheckCorrectStructure(dockerFileDirectory)
	if err != nil {
		return err
	}

	utils.LogMessage("Dockerfile: %s\r\n", dockerFile)

	dockerVersion, err := FetchVersionFromDockerFile(dockerFile)
	if err != nil {
		return err
	}
	utils.LogMessage("\tName: %s\r\n", utils.GlobalOptions.Name)
	utils.LogMessage("\tVersion: %s\r\n", dockerVersion)

	vers, err := ConditionallyBuild(dockerVersion, dockerFile)
	if err != nil {
		return err
	}

	if len(vers.Version) == 0 {
		utils.QuitSuccess()
	}

	var containerConf ContainerConfig

	attributesPath := fmt.Sprintf("%s/attributes.yml", dockerFileDirectory)
	if utils.PathExists(attributesPath) {
		recConf, err := ParseContainerConfig(attributesPath)
		if err != nil {
			return err
		} else {
			utils.LogMessage("Using %s\r\n", attributesPath)
		}
		containerConf = recConf
	} else {
		containerConf = ContainerConfig{}
	}

	containerConf.Image = vers

	err = SpinUpContainer(containerConf)

	if err != nil {
		return err
	}

	err = RemoveImagesNotAt(vers.Version, vers.Image.ID)

	if err != nil {
		return err
	}

	return nil
}
