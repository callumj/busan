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
		utils.LogMessage("Image version: %s\r\n", ver.Version)
		uploaded, err := UploadImage(version, dockerFile)
		if err != nil {
			return VersionImage{}, err
		}

		utils.LogMessage("\tUploaded: %s\r\n", uploaded.Image.ID)
		return uploaded, nil
	} else {
		utils.LogMessage("Server has current version (%s)\r\n", ver.Version)
	}

	return VersionImage{}, nil
}

func GetCurrentVersion() (VersionImage, error) {
	allImages, err := remote.DockerClient.ListImages(true)
	if err != nil {
		return VersionImage{}, err
	}

	for _, img := range allImages {
		for _, tag := range img.RepoTags {
			if strings.Index(tag, utils.GlobalOptions.Name) == 0 {
				parts := strings.Split(tag, ":")
				if len(parts) >= 2 {
					ver := parts[1]
					extracted := versionGrab.FindStringSubmatch(ver)
					if len(extracted) >= 2 {
						return VersionImage{
							Image:   img,
							Version: extracted[1],
						}, nil
					}
				}
			}
		}
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

	outBuf.WriteString("\r\n")
	outBuf.Flush()
	if err := remote.DockerClient.BuildImage(opts); err != nil {
		return VersionImage{}, err
	}

	cur, err := GetCurrentVersion()

	if err != nil {
		return VersionImage{}, err
	} else {
		return cur, nil
	}
}
