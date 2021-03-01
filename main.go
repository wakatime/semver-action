package main

import (
	"fmt"
	"os"

	"github.com/wakatime/semver-action/cmd/generate"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
)

func main() {
	log.SetHandler(cli.Default)

	result, err := generate.Run()
	if err != nil {
		log.Errorf("failed to generate semver version: %s\n", err)

		os.Exit(1)
	}

	// Print previous tag.
	log.Infof("PREVIOUS_TAG: %s", result.PreviousTag)
	fmt.Printf("::set-output name=PREVIOUS_TAG::%s\n", result.PreviousTag)

	// Print calculated semver tag.
	log.Infof("SEMVER_TAG: %s", result.SemverTag)
	fmt.Printf("::set-output name=SEMVER_TAG::%s\n", result.SemverTag)

	// Print is prerelease.
	log.Infof("IS_PRERELEASE: %v", result.IsPrerelease)
	fmt.Printf("::set-output name=IS_PRERELEASE::%v\n", result.IsPrerelease)

	os.Exit(0)
}
