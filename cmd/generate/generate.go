package generate

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/wakatime/semver-action/pkg/git"

	"github.com/apex/log"
	"github.com/blang/semver/v4"
)

// nolint: gochecknoglobals
var (
	branchBugfixPrefixRegex  = regexp.MustCompile(`(?i)^bugfix/.+`)
	branchDocPrefixRegex     = regexp.MustCompile(`(?i)^docs?/.+`)
	branchFeaturePrefixRegex = regexp.MustCompile(`(?i)^feature/.+`)
	branchHotfixPrefixRegex  = regexp.MustCompile(`(?i)^hotfix/.+`)
	branchMajorPrefixRegex   = regexp.MustCompile(`(?i)^major/.+`)
	branchMiscPrefixRegex    = regexp.MustCompile(`(?i)^misc/.+`)
	branchResyncPrefixRegex  = regexp.MustCompile(`(?i)^resync/.+`)
)

const tagDefault = "0.0.0"

type gitClient interface {
	CurrentBranch() (string, error)
	IsRepo() bool
	LatestTag() string
	AncestorTag(include, exclude, branch string) string
	SourceBranch(commitHash string) (string, error)
}

// Result contains the result of Run().
type Result struct {
	PreviousTag  string
	AncestorTag  string
	SemverTag    string
	IsPrerelease bool
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

	gc := git.NewGit(params.RepoDir)

	return Tag(params, gc)
}

// Tag returns the calculated semantica version.
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

	var tag *semver.Version

	latestTag := gc.LatestTag()
	if latestTag == "" {
		tag, _ = semver.New(tagDefault)
	} else {
		parsed, err := semver.ParseTolerant(latestTag)
		if err != nil {
			return Result{}, fmt.Errorf("failed to parse tag %q or not valid semantic version: %s", latestTag, err)
		}
		tag = &parsed
	}

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

	// If branch is prefixed with doc or misc and the latest tag is equal to the
	// ancestor develop tag excluding prerelease part, then it will use ancestor one instead.
	if (branchDocPrefixRegex.MatchString(source) || branchMiscPrefixRegex.MatchString(source)) &&
		dest == params.DevelopBranchName {
		ancestorDevelopTag := gc.AncestorTag(
			fmt.Sprintf("%s[0-9]*-%s*", params.Prefix, params.PrereleaseID),
			"",
			dest)

		parsed, err := semver.ParseTolerant(ancestorDevelopTag)
		if err != nil {
			return Result{}, fmt.Errorf("failed to parse tag %q or not valid semantic version: %s", latestTag, err)
		}

		if tag.String() == parsed.FinalizeVersion() {
			tag = &parsed
		}
	}

	var (
		finalTag       string
		ancestorTag    string
		includePattern string
		excludePattern string
		isPrerelease   bool
	)

	switch method {
	case "build":
		{
			isPrerelease = true
			includePattern = fmt.Sprintf("%s[0-9]*-%s*", params.Prefix, params.PrereleaseID)

			buildNumber, _ := semver.NewPRVersion("0")

			if len(tag.Pre) > 1 && version == "" {
				buildNumber = tag.Pre[1]
			}

			tag.Pre = nil

			preVersion, err := semver.NewPRVersion(params.PrereleaseID)
			if err != nil {
				return Result{}, fmt.Errorf("failed to create new prerelease version: %s", err)
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
		if len(tag.Pre) > 0 {
			isPrerelease = true
			includePattern = fmt.Sprintf("%s[0-9]*-%s*", params.Prefix, params.PrereleaseID)
		} else {
			includePattern = fmt.Sprintf("%s[0-9]*", params.Prefix)
			excludePattern = fmt.Sprintf("%s[0-9]*-%s*", params.Prefix, params.PrereleaseID)
		}

		finalTag = params.Prefix + tag.String()
	default:
		includePattern = fmt.Sprintf("%s[0-9]*", params.Prefix)
		excludePattern = fmt.Sprintf("%s[0-9]*-%s*", params.Prefix, params.PrereleaseID)
		finalTag = params.Prefix + tag.FinalizeVersion()
	}

	ancestorTag = gc.AncestorTag(includePattern, excludePattern, dest)

	return Result{
		PreviousTag:  previousTag,
		AncestorTag:  ancestorTag,
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

	// doc into develop branch
	if branchDocPrefixRegex.MatchString(sourceBranch) && destBranch == developBranchName {
		return "build", ""
	}

	// feature into develop
	if branchFeaturePrefixRegex.MatchString(sourceBranch) && destBranch == developBranchName {
		return "build", "minor"
	}

	// major into develop
	if branchMajorPrefixRegex.MatchString(sourceBranch) && destBranch == developBranchName {
		return "build", "major"
	}

	// misc into develop branch
	if branchMiscPrefixRegex.MatchString(sourceBranch) && destBranch == developBranchName {
		return "build", ""
	}

	// hotfix into main branch
	if branchHotfixPrefixRegex.MatchString(sourceBranch) && destBranch == mainBranchName {
		return "hotfix", ""
	}

	// resync into develop
	if branchResyncPrefixRegex.MatchString(sourceBranch) && destBranch == developBranchName {
		return "build", "patch"
	}

	// develop branch into main branch
	if sourceBranch == developBranchName && destBranch == mainBranchName {
		return "final", ""
	}

	return "build", ""
}
