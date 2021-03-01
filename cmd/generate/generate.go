package generate

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/wakatime/semver-action/pkg/git"

	"github.com/apex/log"
	"github.com/blang/semver/v4"
)

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

// Run generates a semantic version using the commit sha.
func Run() (Result, error) {
	p, err := LoadParams()
	if err != nil {
		return Result{}, fmt.Errorf("failed to load parameters: %s", err)
	}

	if p.Debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("debug logs enabled\n")
	}

	log.Debug(p.String())

	if !git.IsRepo() {
		return Result{}, fmt.Errorf("current folder is not a git repository")
	}

	tagSource := "git"

	if p.BaseVersion != nil {
		tagSource = "parameter"
	}

	source, dest, err := getBranchesFromCommit(p.CommitSha, p.RepoDir)
	if err != nil {
		return Result{}, fmt.Errorf("failed extracting source and dest branches from commit: %s", err)
	}

	log.Debugf("source branch: %q, dest branch: %q\n", source, dest)

	method, version := determineBumpStrategy(p.Bump, source, dest, p.MainBranchName, p.DevelopBranchName)

	log.Debugf("method: %q, version: %q", method, version)

	tag, err := getLatestTagOrDefault(p.Prefix, p.RepoDir)
	if err != nil {
		return Result{}, fmt.Errorf("failed getting latest tag: %s", err)
	}

	previousTag := p.Prefix + tag.String()

	if tagSource != "git" {
		tag = p.BaseVersion
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

			preVersion, err := semver.NewPRVersion(p.PrereleaseID)
			if err != nil {
				return Result{}, fmt.Errorf("failed to create new pre-release version: %s", err)
			}

			tag.Pre = append(tag.Pre, preVersion)

			buildVersion, err := semver.NewPRVersion(strconv.Itoa(int(buildNumber.VersionNum + 1)))
			if err != nil {
				return Result{}, fmt.Errorf("failed to create new build version: %s", err)
			}

			tag.Pre = append(tag.Pre, buildVersion)

			finalTag = p.Prefix + tag.String()
		}
	case "major", "minor", "patch":
		isPrerelease = len(tag.Pre) > 0

		finalTag = p.Prefix + tag.String()
	default:
		finalTag = p.Prefix + tag.FinalizeVersion()
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

func getBranchesFromCommit(hash string, repoDir string) (string, string, error) {
	dest, err := git.Clean(git.Run("-C", repoDir, "rev-parse", "--abbrev-ref", "HEAD", "--quiet"))
	if err != nil {
		return "", "", fmt.Errorf("could not get current branch: %s", err)
	}

	message, err := git.Clean(git.Run("-C", repoDir, "log", "-1", "--pretty=%B", hash))
	if err != nil {
		return "", "", fmt.Errorf("could not get message from commit: %s", err)
	}

	source, err := getSourceBranchFromCommit(message)
	if err != nil {
		return "", "", fmt.Errorf("could not parse source branch: %s", err)
	}

	return source, dest, nil
}

func getSourceBranchFromCommit(message string) (string, error) {
	match := mergePRRegex.FindStringSubmatch(message)

	paramsMap := make(map[string]string)
	for i, name := range mergePRRegex.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}

	if len(paramsMap) == 0 || paramsMap["source"] == "" {
		return "", errors.New("no source branch found")
	}

	splitted := strings.SplitN(paramsMap["source"], "/", 2)

	if len(splitted) < 2 {
		return "", fmt.Errorf("commit message does not contain expected format: %s", paramsMap["source"])
	}

	return splitted[1], nil
}

func getLatestTagOrDefault(prefix string, repoDir string) (*semver.Version, error) {
	var (
		prefixRe = regexp.MustCompile(fmt.Sprintf("^%s", prefix))
		tag      *semver.Version
		err      error
	)

	for _, fn := range []func() (string, error){
		func() (string, error) {
			return git.Clean(git.Run("-C", repoDir, "tag", "--points-at", "HEAD", "--sort", "-version:creatordate"))
		},
		func() (string, error) {
			return git.Clean(git.Run("-C", repoDir, "describe", "--tags", "--abbrev=0"))
		},
		func() (string, error) {
			return "0.0.0", nil
		},
	} {
		tagStr, _ := fn()
		if tagStr != "" {
			tagStr = prefixRe.ReplaceAllLiteralString(tagStr, "")
			parsed, err := semver.Parse(tagStr)
			if err != nil {
				log.Errorf("failed to parse tag %q or not valid semantic version: %s", tagStr, err)
			}
			return &parsed, nil
		}
	}

	return tag, err
}
