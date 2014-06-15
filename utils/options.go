package utils

type OptionSet struct {
	HostAddress string `short:"d" long:"docker-host" description:"unix:// or tcp:// address to Docker host"`
	Name        string `short:"n" long:"name" description:"The name of the configurations"`
	Version     bool   `short:"v" long:"version" description:"Print the version"`
}

var GlobalOptions OptionSet
