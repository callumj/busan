package main

import (
	"github.com/callumj/docker-mate/app"
	"os"
)

func main() {
	args := os.Args
	app.Run(args)
}
