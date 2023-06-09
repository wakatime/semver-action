package main

import (
	"fmt"
	"os"

	"github.com/wakatime/semver-action/cmd/generate"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/gofrs/uuid"
)

func main() {
	log.SetHandler(cli.Default)

	result, err := generate.Run()
	if err != nil {
		log.Errorf("failed to generate semver version: %s\n", err)

		os.Exit(1)
	}

	outputFilepath := os.Getenv("GITHUB_OUTPUT")

	// Print previous tag.
	log.Infof("PREVIOUS_TAG: %s", result.PreviousTag)

	if err := setOutput(outputFilepath, "PREVIOUS_TAG", result.PreviousTag); err != nil {
		log.Errorf("%s\n", err)

		os.Exit(1)
	}

	// Print ancestor tag.
	log.Infof("ANCESTOR_TAG: %s", result.AncestorTag)

	if err := setOutput(outputFilepath, "ANCESTOR_TAG", result.AncestorTag); err != nil {
		log.Errorf("%s\n", err)

		os.Exit(1)
	}

	// Print calculated semver tag.
	log.Infof("SEMVER_TAG: %s", result.SemverTag)

	if err := setOutput(outputFilepath, "SEMVER_TAG", result.SemverTag); err != nil {
		log.Errorf("%s\n", err)

		os.Exit(1)
	}

	// Print is prerelease.
	log.Infof("IS_PRERELEASE: %v", result.IsPrerelease)

	if err := setOutput(outputFilepath, "IS_PRERELEASE", fmt.Sprintf("%v", result.IsPrerelease)); err != nil {
		log.Errorf("%s\n", err)

		os.Exit(1)
	}
}

func setOutput(fp, key, value string) error {
	f, err := os.OpenFile(fp, os.O_APPEND|os.O_WRONLY, 0600) // nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to open github output file: %s", err)
	}

	defer func() {
		_ = f.Close()
	}()

	delimiter := fmt.Sprintf("ghadelimiter_%s", newId())

	if _, err := f.WriteString(fmt.Sprintf("%s<<%s\n%v\n%s\n", key, delimiter, value, delimiter)); err != nil {
		return fmt.Errorf("failed to write %s to output: %s", key, err)
	}

	return nil
}

func newId() string {
	id, err := uuid.NewV4()
	if err != nil {
		log.Errorf("failed to generate delimier uuid: %s\n", err)

		os.Exit(1)
	}

	return id.String()
}
