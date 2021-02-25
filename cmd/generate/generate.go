package generate

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	ghtag "github.com/wakatime/semver-action/cmd/tag"
	"github.com/wakatime/semver-action/pkg/exitcode"
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

// Run generates a semantic version using the commit sha.
func Run() {
	p, err := LoadParams()
	if err != nil {
		log.Fatalf("failed to load parameters: %s\n", err)

		os.Exit(exitcode.ErrDefault)
	}

	if p.Debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("debug logs enabled\n")
	}

	log.Debug(p.String())

	if !git.IsRepo() {
		log.Fatal("current folder is not a git repository")

		os.Exit(exitcode.ErrDefault)
	}

	tagSource := "git"

	if p.BaseVersion != nil {
		tagSource = "parameter"
	}

	source, dest, err := getBranchesFromCommit(p.CommitSha)
	if err != nil {
		log.Errorf("failed extracting source and dest branches from commit: %s", err)

		os.Exit(exitcode.ErrDefault)
	}

	log.Debugf("source branch: %q, dest branch: %q\n", source, dest)

	method, version := defineBumpStrategy(p.Bump, source, dest, p.MainBranchName, p.DevelopBranchName)

	log.Debugf("method: %q, version: %q", method, version)

	tag, err := getLatestTagOrDefault(p.Prefix)
	if err != nil {
		log.Errorf("failed getting latest tag: %s", err)
	}

	// Print current tag.
	fmt.Printf("::set-output name=PREVIOUS_TAG::%s\n", p.Prefix+tag.String())

	if tagSource != "git" {
		tag = p.BaseVersion
	}

	if tagSource == "git" && version == "major" && method == "build" {
		log.Debug("incrementing major")
		if err := tag.IncrementMajor(); err != nil {
			log.Errorf("failed to increment major version: %s", err)

			os.Exit(exitcode.ErrDefault)
		}
	}

	if tagSource == "git" && version == "minor" && method == "build" {
		log.Debug("incrementing minor")
		if err := tag.IncrementMinor(); err != nil {
			log.Errorf("failed to increment minor version: %s", err)

			os.Exit(exitcode.ErrDefault)
		}
	}

	if tagSource == "git" && (version == "patch" && method == "build") || method == "hotfix" {
		log.Debug("incrementing patch")
		if err := tag.IncrementPatch(); err != nil {
			log.Errorf("failed to increment patch version: %s", err)

			os.Exit(exitcode.ErrDefault)
		}
	}

	var finalTag string

	if method == "build" {
		buildNumber, _ := semver.NewPRVersion("0")

		if len(tag.Pre) > 1 && version == "" {
			buildNumber = tag.Pre[1]
		}

		tag.Pre = nil

		preVersion, err := semver.NewPRVersion(p.PrereleaseID)
		if err != nil {
			log.Errorf("failed to create new pre-release version: %s", err)

			os.Exit(exitcode.ErrDefault)
		}

		tag.Pre = append(tag.Pre, preVersion)

		buildVersion, err := semver.NewPRVersion(strconv.Itoa(int(buildNumber.VersionNum + 1)))
		if err != nil {
			log.Errorf("failed to create new build version: %s", err)

			os.Exit(exitcode.ErrDefault)
		}

		tag.Pre = append(tag.Pre, buildVersion)

		finalTag = p.Prefix + tag.String()
	}

	if method == "hotfix" || method == "final" {
		finalTag = p.Prefix + tag.FinalizeVersion()
	}

	// Create tag locally and push it.
	if !p.DryRun {
		err := ghtag.Create(ghtag.Params{
			CommitSha:  p.CommitSha,
			Client:     p.Client,
			Owner:      p.Owner,
			Repository: p.Repository,
			TagMessage: p.TagMessage,
			Tag:        finalTag,
		})
		if err != nil {
			log.Error(err.Error())

			os.Exit(exitcode.ErrDefault)
		}
	}

	fmt.Printf("::set-output name=SEMVER_TAG::%s\n", finalTag)

	os.Exit(exitcode.Success)
}

// defineBumpStrategy defines the strategy for semver to bump product version.
func defineBumpStrategy(bump, sourceBranch, destBranch, mainBranchName, developBranchName string) (string, string) {
	if bump == "auto" {
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

	return bump, ""
}

func getBranchesFromCommit(hash string) (string, string, error) {
	dest, err := git.Clean(git.Run("rev-parse", "--abbrev-ref", "HEAD", "--quiet"))
	if err != nil {
		return "", "", fmt.Errorf("could not get current branch: %s", err)
	}

	message, err := git.Clean(git.Run("log", "-1", "--pretty=%B", hash))
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

func getLatestTagOrDefault(prefix string) (*semver.Version, error) {
	var (
		prefixRe = regexp.MustCompile(fmt.Sprintf("^%s", prefix))
		tag      *semver.Version
		err      error
	)

	for _, fn := range []func() (string, error){
		func() (string, error) {
			return git.Clean(git.Run("tag", "--points-at", "HEAD", "--sort", "-version:creatordate"))
		},
		func() (string, error) {
			return git.Clean(git.Run("describe", "--tags", "--abbrev=0"))
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
