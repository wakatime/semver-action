package main

import (
	"github.com/wakatime/semver-action/cmd/generate"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
)

func main() {
	log.SetHandler(cli.Default)

	generate.Run()
}
