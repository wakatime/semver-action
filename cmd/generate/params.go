package generate

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/wakatime/semver-action/pkg/actions"

	"github.com/blang/semver/v4"
)

var (
	//nolint
	commitShaRegex = regexp.MustCompile(`\b[0-9a-f]{5,40}\b`)
	// nolint
	validBumpStrategies = []string{"auto", "major", "minor", "patch"}
)

// Params contains semver generate command parameters.
type Params struct {
	CommitSha         string
	RepoDir           string
	Bump              string
	BaseVersion       *semver.Version
	Prefix            string
	PrereleaseID      string
	MainBranchName    string
	DevelopBranchName string
	Debug             bool
}

// LoadParams loads semver generate config params.
func LoadParams() (Params, error) {
	var commitSha string

	if commitShaStr := os.Getenv("GITHUB_SHA"); commitShaStr != "" {
		if !commitShaRegex.MatchString(commitShaStr) {
			return Params{}, fmt.Errorf("invalid commit-sha format: %s", commitShaStr)
		}

		commitSha = commitShaStr
	}

	var repoDir string = "."

	if repoDirStr := actions.GetInput("repo_dir"); repoDirStr != "" {
		repoDir = repoDirStr
	}

	var bump string = "auto"

	if bumpStr := actions.GetInput("bump"); bumpStr != "" {
		if !stringInSlice(bumpStr, validBumpStrategies) {
			return Params{}, fmt.Errorf("invalid bump value: %s", bumpStr)
		}

		bump = bumpStr
	}

	var debug bool = true

	if debugStr := actions.GetInput("debug"); debugStr != "" {
		parsed, err := strconv.ParseBool(debugStr)
		if err != nil {
			return Params{}, fmt.Errorf("invalid debug argument: %s", debugStr)
		}

		debug = parsed
	}

	var prefix string = "v"

	if prefixStr := actions.GetInput("prefix"); prefixStr != "" {
		prefix = prefixStr
	}

	var baseVersion *semver.Version

	if baseVersionStr := actions.GetInput("base_version"); baseVersionStr != "" {
		prefixRe := regexp.MustCompile(fmt.Sprintf("^%s", prefix))
		baseVersionStr = prefixRe.ReplaceAllLiteralString(baseVersionStr, "")

		parsed, err := semver.Parse(baseVersionStr)
		if err != nil {
			return Params{}, fmt.Errorf("invalid base_version format: %s", baseVersionStr)
		}

		baseVersion = &parsed
	}

	var mainBranchName string = "master"

	if mainBranchNameStr := actions.GetInput("main_branch_name"); mainBranchNameStr != "" {
		mainBranchName = mainBranchNameStr
	}

	var developBranchName string = "develop"

	if developBranchNameStr := actions.GetInput("develop_branch_name"); developBranchNameStr != "" {
		developBranchName = developBranchNameStr
	}

	var prereleaseID string = "pre"

	if prereleaseIDStr := actions.GetInput("prerelease_id"); prereleaseIDStr != "" {
		prereleaseID = prereleaseIDStr
	}

	return Params{
		CommitSha:         commitSha,
		RepoDir:           repoDir,
		Bump:              bump,
		BaseVersion:       baseVersion,
		Prefix:            prefix,
		PrereleaseID:      prereleaseID,
		MainBranchName:    mainBranchName,
		DevelopBranchName: developBranchName,
		Debug:             debug,
	}, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}

	return false
}

func (p Params) String() string {
	var baseVersion string
	if p.BaseVersion != nil {
		baseVersion = p.BaseVersion.String()
	}

	return fmt.Sprintf(
		"commit sha: %q, bump: %q, base version: %q, prefix: %q,"+
			" prerelease id: %q, main branch name: %q, develop branch name: %q,"+
			" repo dir: %q, debug: %t\n",
		p.CommitSha,
		p.Bump,
		baseVersion,
		p.Prefix,
		p.PrereleaseID,
		p.MainBranchName,
		p.DevelopBranchName,
		p.RepoDir,
		p.Debug,
	)
}
