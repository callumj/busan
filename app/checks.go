package app

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/callumj/busan/utils"
	"os"
	"path"
	"regexp"
)

var versionReg = regexp.MustCompile(`^#\s*VERSION\s+((?:\d+|\.)+)`)

func CheckCorrectStructure(dockerFileDirectory string) (string, error) {
	statInfo, err := os.Stat(dockerFileDirectory)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("Path %s does not exist", dockerFileDirectory)
		}
		return "", err
	}

	if !statInfo.IsDir() {
		return "", fmt.Errorf("%s is not a directory", dockerFileDirectory)
	}

	// setup the name if needed
	if len(utils.GlobalOptions.Name) == 0 {
		utils.GlobalOptions.Name = path.Base(dockerFileDirectory)
	}

	// test for Dockerfile existence
	dockerFile := fmt.Sprintf("%s/Dockerfile", dockerFileDirectory)
	statInfo, err = os.Stat(dockerFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("Dockerfile %s does not exist", dockerFile)
		}
		return "", err
	}

	if statInfo.IsDir() {
		return "", fmt.Errorf("%v is a directory", dockerFile)
	}

	return dockerFile, nil
}

func FetchVersionFromDockerFile(dockerFile string) (string, error) {
	file, err := os.Open(dockerFile)
	defer file.Close()
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := scanner.Text()
		if versionReg.MatchString(str) {
			extracted := versionReg.FindStringSubmatch(str)
			if len(extracted) >= 2 {
				return extracted[1], nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", errors.New("Unable to fetch VERSION comment")
}
