package generate

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/wakatime/semver-action/pkg/git"

	"github.com/apex/log"
	"github.com/blang/semver/v4"
)

const tagDefault = "0.0.0"

var (
	// nolint
	branchHotfixPrefixRegex = regexp.MustCompile(`(?i)^hotfix(es)?/.*`)
	// nolint
	branchFeaturePrefixRegex = regexp.MustCompile(`(?i)^feature(s)?/.*`)
	// nolint
	branchBugfixPrefixRegex = regexp.MustCompile(`(?i)^bugfix(es)?/.*`)
	// nolint
	branchMajorPrefixRegex = regexp.MustCompile(`(?i)^major/.*`)
	// nolint
	mergePRRegex = regexp.MustCompile(`Merge pull request #([0-9])+ from (?P<source>.*)+`)
)

// Result contains the result of Run().
type Result struct {
	PreviousTag  string
	SemverTag    string
	IsPrerelease bool
}

type gitClient interface {
	CurrentBranch() (string, error)
	IsRepo() bool
	LatestTag() (string, error)
	SourceBranch(commitHash string) (string, error)
}

// Run generates a semantic version using the commit sha.
func Run() (Result, error) {
	params, err := LoadParams()
	if err != nil {
		return Result{}, fmt.Errorf("failed to load parameters: %s", err)
	}

	if params.Debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("debug logs enabled\n")
	}

	log.Debug(params.String())

	gc := git.NewClient(params.RepoDir)
	return Tag(params, gc)
}

func Tag(params Params, gc gitClient) (Result, error) {
	if !gc.IsRepo() {
		return Result{}, fmt.Errorf("current folder is not a git repository")
	}

	tagSource := "git"

	if params.BaseVersion != nil {
		tagSource = "parameter"
	}

	dest, err := gc.CurrentBranch()
	if err != nil {
		return Result{}, fmt.Errorf("failed to extract dest branche from commit: %s", err)
	}

	log.Debugf("dest branch: %q\n", dest)

	source, err := gc.SourceBranch(params.CommitSha)
	if err != nil {
		return Result{}, fmt.Errorf("failed to extract source branch from commit: %s", err)
	}

	log.Debugf("source branch: %q\n", source)

	method, version := determineBumpStrategy(params.Bump, source, dest, params.MainBranchName, params.DevelopBranchName)

	log.Debugf("method: %q, version: %q", method, version)

	tagStr, _ := gc.LatestTag()
	if tagStr == "" {
		tagStr = tagDefault
	}

	tagStr = strings.TrimPrefix(tagStr, params.Prefix)
	tagPtr, err := semver.Parse(tagStr)
	if err != nil {
		return Result{}, fmt.Errorf("failed to parse tag %q or not valid semantic version: %s", tagStr, err)
	}

	tag := &tagPtr

	previousTag := params.Prefix + tag.String()

	if tagSource != "git" {
		tag = params.BaseVersion
	}

	if (version == "major" && method == "build") || method == "major" {
		log.Debug("incrementing major")
		if err := tag.IncrementMajor(); err != nil {
			return Result{}, fmt.Errorf("failed to increment major version: %s", err)
		}
	}

	if (version == "minor" && method == "build") || method == "minor" {
		log.Debug("incrementing minor")
		if err := tag.IncrementMinor(); err != nil {
			return Result{}, fmt.Errorf("failed to increment minor version: %s", err)
		}
	}

	if (version == "patch" && method == "build") || method == "patch" || method == "hotfix" {
		log.Debug("incrementing patch")
		if err := tag.IncrementPatch(); err != nil {
			return Result{}, fmt.Errorf("failed to increment patch version: %s", err)
		}
	}

	var (
		finalTag     string
		isPrerelease bool
	)

	switch method {
	case "build":
		{
			isPrerelease = true

			buildNumber, _ := semver.NewPRVersion("0")

			if len(tag.Pre) > 1 && version == "" {
				buildNumber = tag.Pre[1]
			}

			tag.Pre = nil

			preVersion, err := semver.NewPRVersion(params.PrereleaseID)
			if err != nil {
				return Result{}, fmt.Errorf("failed to create new pre-release version: %s", err)
			}

			tag.Pre = append(tag.Pre, preVersion)

			buildVersion, err := semver.NewPRVersion(strconv.Itoa(int(buildNumber.VersionNum + 1)))
			if err != nil {
				return Result{}, fmt.Errorf("failed to create new build version: %s", err)
			}

			tag.Pre = append(tag.Pre, buildVersion)

			finalTag = params.Prefix + tag.String()
		}
	case "major", "minor", "patch":
		isPrerelease = len(tag.Pre) > 0

		finalTag = params.Prefix + tag.String()
	default:
		finalTag = params.Prefix + tag.FinalizeVersion()
	}

	return Result{
		PreviousTag:  previousTag,
		SemverTag:    finalTag,
		IsPrerelease: isPrerelease,
	}, nil
}

// determineBumpStrategy determines the strategy for semver to bump product version.
func determineBumpStrategy(bump, sourceBranch, destBranch, mainBranchName, developBranchName string) (string, string) {
	if bump != "auto" {
		return bump, ""
	}

	// bugfix into develop branch
	if branchBugfixPrefixRegex.MatchString(sourceBranch) && destBranch == developBranchName {
		return "build", "patch"
	}

	// feature into develop
	if branchFeaturePrefixRegex.MatchString(sourceBranch) && destBranch == developBranchName {
		return "build", "minor"
	}

	// major into develop
	if branchMajorPrefixRegex.MatchString(sourceBranch) && destBranch == developBranchName {
		return "build", "major"
	}

	// hotfix into main branch
	if branchHotfixPrefixRegex.MatchString(sourceBranch) && destBranch == mainBranchName {
		return "hotfix", ""
	}

	// develop branch into main branch
	if sourceBranch == developBranchName && destBranch == mainBranchName {
		return "final", ""
	}

	return "build", ""
}
