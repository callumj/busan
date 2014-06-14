package app

import (
	"archive/tar"
	"bufio"
	"bytes"
	"fmt"
	"github.com/callumj/docker-mate/remote"
	"github.com/callumj/docker-mate/utils"
	"github.com/fsouza/go-dockerclient"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var versionGrab = regexp.MustCompile(`v((?:\d|\.)+)`)

type VersionImage struct {
	Image   docker.APIImages
	Version string
}

func ConditionallyBuild(version, dockerFile string) (VersionImage, error) {
	ver, err := GetCurrentVersion()
	if err != nil {
		return VersionImage{}, err
	}

	if ver.Version != version {
		if len(ver.Version) != 0 {
			utils.LogMessage("Image version: %s\r\n", ver.Version)
		} else {
			utils.LogMessage("Building version: %s\r\n", version)
		}
		uploaded, err := UploadImage(version, dockerFile)
		if err != nil {
			return VersionImage{}, err
		}

		utils.LogMessage("Uploaded: %s\r\n", uploaded.Image.ID)
		return uploaded, nil
	} else {
		utils.LogMessage("Server has current version (%s)\r\n", ver.Version)
	}

	return ver, nil
}

func GetCurrentVersion() (VersionImage, error) {
	var received *VersionImage
	err := loopOnFoundImages(func(ver string, img docker.APIImages) {
		received = &VersionImage{
			Image:   img,
			Version: ver,
		}
	})

	if err != nil {
		return VersionImage{}, err
	}

	if received != nil {
		return *received, nil
	}

	return VersionImage{}, nil
}

func UploadImage(version, dockerFile string) (VersionImage, error) {
	directory := path.Dir(dockerFile)

	tarBuf := new(bytes.Buffer)
	tw := tar.NewWriter(tarBuf)

	walkFn := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			hdr := &tar.Header{
				Name: info.Name(),
				Size: info.Size(),
			}
			tw.WriteHeader(hdr)

			filePntr, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("Unable to read in %s (%v)\r\n", info.Name(), err)
			}
			defer filePntr.Close()

			// read in chunks for memory
			buf := make([]byte, 1024)
			for {
				// read a chunk
				n, err := filePntr.Read(buf)
				if err != nil && err != io.EOF {
					return fmt.Errorf("Unable to read chunk %s (%v)\r\n", info.Name(), err)
				}
				if n == 0 {
					break
				}

				// write a chunk
				if _, err := tw.Write(buf[:n]); err != nil {
					return fmt.Errorf("Unable to write chunk %s (%v)\r\n", info.Name(), err)
				}
			}
		}

		return nil
	}
	err := filepath.Walk(directory, walkFn)

	if err != nil {
		return VersionImage{}, err
	}

	tw.Close()

	outBuf := bufio.NewWriter(os.Stdout)
	opts := docker.BuildImageOptions{
		Name:         utils.BuildName(version),
		InputStream:  tarBuf,
		OutputStream: outBuf,
	}

	if err := remote.DockerClient.BuildImage(opts); err != nil {
		return VersionImage{}, err
	}

	outBuf.WriteString("\033[0m")
	outBuf.WriteString("\r\n")
	outBuf.Flush()

	cur, err := GetCurrentVersion()

	if err != nil {
		return VersionImage{}, err
	} else {
		return cur, nil
	}
}

func RemoveImagesNotAt(version, imgId string) error {
	err := loopOnFoundImages(func(ver string, img docker.APIImages) {
		if ver != version && img.ID != imgId {
			utils.LogMessage("Removing %v\r\n", img.Tag)
			remote.DockerClient.RemoveImage(img.ID)
		} else if ver != version && img.ID == imgId {
			utils.LogMessage("Will not delete %v as the ID is the same, please clean manually\r\n", img.RepoTags)
		}
	})

	return err
}

type onImageFind func(string, docker.APIImages)

func loopOnFoundImages(callback onImageFind) error {
	allImages, err := remote.DockerClient.ListImages(true)
	if err != nil {
		return err
	}

	for _, img := range allImages {
		for _, tag := range img.RepoTags {
			if strings.Index(tag, utils.GlobalOptions.Name) == 0 {
				parts := strings.Split(tag, ":")
				if len(parts) >= 2 {
					ver := parts[1]
					extracted := versionGrab.FindStringSubmatch(ver)
					if len(extracted) >= 2 {
						callback(extracted[1], img)
					}
				}
			}
		}
	}

	return nil
}
