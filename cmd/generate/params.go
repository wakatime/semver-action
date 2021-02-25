package generate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/wakatime/semver-action/pkg/actions"

	"github.com/blang/semver/v4"
	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

var (
	//nolint
	commitShaRegex = regexp.MustCompile(`\b[0-9a-f]{5,40}\b`)
	// nolint
	validBumpStrategies = []string{"auto", "major", "minor", "patch"}
	// nolint
	authTokenRegex = regexp.MustCompile(`[0-9a-fA-F]{40}`)
)

// Params contains semver generate command parameters.
type Params struct {
	CommitSha         string
	Bump              string
	AuthToken         string
	Client            *github.Client
	BaseVersion       *semver.Version
	Prefix            string
	PrereleaseID      string
	MainBranchName    string
	DevelopBranchName string
	TagMessage        string
	Owner             string
	Repository        string
	DryRun            bool
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

	var bump string = "auto"

	if bumpStr := actions.GetInput("bump"); bumpStr != "" {
		if !stringInSlice(bumpStr, validBumpStrategies) {
			return Params{}, fmt.Errorf("invalid bump value: %s", bumpStr)
		}

		bump = bumpStr
	}

	var (
		owner      string
		repository string
	)

	if githubRepositoryStr := os.Getenv("GITHUB_REPOSITORY"); githubRepositoryStr != "" {
		splitted := strings.SplitN(githubRepositoryStr, "/", 2)

		owner = splitted[0]
		repository = splitted[1]
	}

	var authToken string

	if authTokenStr := actions.GetInput("auth_token"); authTokenStr != "" {
		if !authTokenRegex.MatchString(authTokenStr) {
			return Params{}, fmt.Errorf("invalid auth_token format: %s", authTokenStr)
		}

		authToken = authTokenStr
	}

	var dryRun bool

	if dryRunStr := actions.GetInput("dry_run"); dryRunStr != "" {
		parsed, err := strconv.ParseBool(dryRunStr)
		if err != nil {
			return Params{}, fmt.Errorf("invalid dry_run argument: %s", dryRunStr)
		}

		dryRun = parsed
	}

	var githubClient *github.Client

	if authToken == "" && !dryRun {
		return Params{}, errors.New("auth_token is required when dry_run is false")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: authToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	githubClient = client

	var debug bool

	if debugStr := actions.GetInput("debug"); debugStr != "" {
		parsed, err := strconv.ParseBool(debugStr)
		if err != nil {
			return Params{}, fmt.Errorf("invalid debug argument: %s", debugStr)
		}

		debug = parsed
	}

	var baseVersion *semver.Version

	if baseVersionStr := actions.GetInput("base_version"); baseVersionStr != "" {
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

	var prefix string = "v"

	if prefixStr := actions.GetInput("prefix"); prefixStr != "" {
		prefix = prefixStr
	}

	var prereleaseID string = "pre"

	if prereleaseIDStr := actions.GetInput("prerelease_id"); prereleaseIDStr != "" {
		prereleaseID = prereleaseIDStr
	}

	var tagMessage string = "auto tag"

	if tagMessageStr := actions.GetInput("tag_message"); tagMessageStr != "" {
		tagMessage = tagMessageStr
	}

	return Params{
		CommitSha:         commitSha,
		Bump:              bump,
		BaseVersion:       baseVersion,
		Prefix:            prefix,
		PrereleaseID:      prereleaseID,
		MainBranchName:    mainBranchName,
		DevelopBranchName: developBranchName,
		TagMessage:        tagMessage,
		AuthToken:         authToken,
		Client:            githubClient,
		Owner:             owner,
		Repository:        repository,
		DryRun:            dryRun,
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
			" prerelease id: %q, tag message: %q, owner: %s, repository: %s,"+
			" dry run: %t, debug: %t\n",
		p.CommitSha,
		p.Bump,
		baseVersion,
		p.Prefix,
		p.PrereleaseID,
		p.TagMessage,
		p.Owner,
		p.Repository,
		p.DryRun,
		p.Debug,
	)
}
