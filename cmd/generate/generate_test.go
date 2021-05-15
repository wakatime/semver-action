package generate_test

import (
	"testing"

	"github.com/alecthomas/assert"
	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/semver-action/cmd/generate"
)

func TestTag(t *testing.T) {
	tests := map[string]struct {
		CurrentBranch string
		LatestTag     string
		AncestorTag   string
		SourceBranch  string
		Params        generate.Params
		Result        generate.Result
	}{
		"no previous tag": {
			CurrentBranch: "develop",
			SourceBranch:  "major/semver-initial",
			Params: generate.Params{
				CommitSha:         "81918ffc",
				Bump:              "auto",
				Prefix:            "v",
				PrereleaseID:      "alpha",
				MainBranchName:    "master",
				DevelopBranchName: "develop",
			},
			Result: generate.Result{
				PreviousTag:  "v0.0.0",
				AncestorTag:  "",
				SemverTag:    "v1.0.0-alpha.1",
				IsPrerelease: true,
			},
		},
		"first non-development tag": {
			CurrentBranch: "master",
			LatestTag:     "1.0.0-alpha.1",
			AncestorTag:   "e63c125b",
			SourceBranch:  "develop",
			Params: generate.Params{
				CommitSha:         "81918ffc",
				Bump:              "auto",
				Prefix:            "v",
				PrereleaseID:      "alpha",
				MainBranchName:    "master",
				DevelopBranchName: "develop",
			},
			Result: generate.Result{
				PreviousTag:  "v1.0.0-alpha.1",
				AncestorTag:  "e63c125b",
				SemverTag:    "v1.0.0",
				IsPrerelease: false,
			},
		},
		"doc branch into develop": {
			CurrentBranch: "develop",
			LatestTag:     "0.2.1-alpha.1",
			SourceBranch:  "doc/semver-initial",
			Params: generate.Params{
				CommitSha:         "81918ffc",
				Bump:              "auto",
				Prefix:            "v",
				PrereleaseID:      "alpha",
				MainBranchName:    "master",
				DevelopBranchName: "develop",
			},
			Result: generate.Result{
				PreviousTag:  "v0.2.1-alpha.1",
				SemverTag:    "v0.2.1-alpha.2",
				IsPrerelease: true,
			},
		},
		"feature branch into develop": {
			CurrentBranch: "develop",
			LatestTag:     "0.2.1",
			SourceBranch:  "feature/semver-initial",
			Params: generate.Params{
				CommitSha:         "81918ffc",
				Bump:              "auto",
				Prefix:            "v",
				PrereleaseID:      "alpha",
				MainBranchName:    "master",
				DevelopBranchName: "develop",
			},
			Result: generate.Result{
				PreviousTag:  "v0.2.1",
				SemverTag:    "v0.3.0-alpha.1",
				IsPrerelease: true,
			},
		},
		"misc branch into develop": {
			CurrentBranch: "develop",
			LatestTag:     "0.2.1-alpha.1",
			SourceBranch:  "misc/semver-initial",
			Params: generate.Params{
				CommitSha:         "81918ffc",
				Bump:              "auto",
				Prefix:            "v",
				PrereleaseID:      "alpha",
				MainBranchName:    "master",
				DevelopBranchName: "develop",
			},
			Result: generate.Result{
				PreviousTag:  "v0.2.1-alpha.1",
				SemverTag:    "v0.2.1-alpha.2",
				IsPrerelease: true,
			},
		},
		"merge develop into master": {
			CurrentBranch: "master",
			LatestTag:     "1.4.17-alpha.1",
			SourceBranch:  "develop",
			Params: generate.Params{
				CommitSha:         "81918ffc",
				Bump:              "auto",
				Prefix:            "v",
				PrereleaseID:      "alpha",
				MainBranchName:    "master",
				DevelopBranchName: "develop",
			},
			Result: generate.Result{
				PreviousTag:  "v1.4.17-alpha.1",
				SemverTag:    "v1.4.17",
				IsPrerelease: false,
			},
		},
		"merge develop into master with previous matching tag": {
			CurrentBranch: "master",
			LatestTag:     "1.4.17-alpha.1",
			AncestorTag:   "v1.4.16",
			SourceBranch:  "develop",
			Params: generate.Params{
				CommitSha:         "81918ffc",
				Bump:              "auto",
				Prefix:            "v",
				PrereleaseID:      "alpha",
				MainBranchName:    "master",
				DevelopBranchName: "develop",
			},
			Result: generate.Result{
				PreviousTag:  "v1.4.17-alpha.1",
				AncestorTag:  "v1.4.16",
				SemverTag:    "v1.4.17",
				IsPrerelease: false,
			},
		},
		"base version set": {
			CurrentBranch: "develop",
			LatestTag:     "2.6.19",
			SourceBranch:  "feature/semver-initial",
			Params: generate.Params{
				CommitSha:         "81918ffc",
				Bump:              "auto",
				BaseVersion:       newSemVerPtr(t, "4.2.0"),
				Prefix:            "v",
				PrereleaseID:      "alpha",
				MainBranchName:    "master",
				DevelopBranchName: "develop",
			},
			Result: generate.Result{
				PreviousTag:  "v2.6.19",
				SemverTag:    "v4.3.0-alpha.1",
				IsPrerelease: true,
			},
		},
		"invalid branch name": {
			CurrentBranch: "develop",
			LatestTag:     "2.6.19-alpha.1",
			SourceBranch:  "semver-initial",
			Params: generate.Params{
				CommitSha:         "81918ffc",
				Bump:              "auto",
				Prefix:            "v",
				PrereleaseID:      "alpha",
				MainBranchName:    "master",
				DevelopBranchName: "develop",
			},
			Result: generate.Result{
				PreviousTag:  "v2.6.19-alpha.1",
				SemverTag:    "v2.6.19-alpha.2",
				IsPrerelease: true,
			},
		},
		"force bump major": {
			CurrentBranch: "develop",
			LatestTag:     "2.6.19-alpha.1",
			SourceBranch:  "semver-initial",
			Params: generate.Params{
				CommitSha:         "81918ffc",
				Bump:              "major",
				Prefix:            "v",
				PrereleaseID:      "alpha",
				MainBranchName:    "master",
				DevelopBranchName: "develop",
			},
			Result: generate.Result{
				PreviousTag:  "v2.6.19-alpha.1",
				SemverTag:    "v3.0.0-alpha.1",
				IsPrerelease: true,
			},
		},
		"force bump minor": {
			CurrentBranch: "develop",
			LatestTag:     "2.6.19-alpha.1",
			SourceBranch:  "semver-initial",
			Params: generate.Params{
				CommitSha:         "81918ffc",
				Bump:              "minor",
				Prefix:            "v",
				PrereleaseID:      "alpha",
				MainBranchName:    "master",
				DevelopBranchName: "develop",
			},
			Result: generate.Result{
				PreviousTag:  "v2.6.19-alpha.1",
				SemverTag:    "v2.7.0-alpha.1",
				IsPrerelease: true,
			},
		},
		"force bump patch": {
			CurrentBranch: "develop",
			LatestTag:     "2.6.19-alpha.1",
			SourceBranch:  "semver-initial",
			Params: generate.Params{
				CommitSha:         "81918ffc",
				Bump:              "patch",
				Prefix:            "v",
				PrereleaseID:      "alpha",
				MainBranchName:    "master",
				DevelopBranchName: "develop",
			},
			Result: generate.Result{
				PreviousTag:  "v2.6.19-alpha.1",
				SemverTag:    "v2.6.20-alpha.1",
				IsPrerelease: true,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gc := initGitClientMock(
				t,
				test.LatestTag,
				test.AncestorTag,
				test.CurrentBranch,
				test.SourceBranch,
				test.Params.CommitSha,
			)

			result, err := generate.Tag(test.Params, gc)
			require.NoError(t, err)

			assert.Equal(t, test.Result, result)
		})
	}
}

func TestTag_IsNotRepo(t *testing.T) {
	gc := &gitClientMock{
		IsRepoFn: func() bool {
			return false
		},
	}

	_, err := generate.Tag(generate.Params{}, gc)
	require.Error(t, err)

	assert.EqualError(t, err, "current folder is not a git repository")
}

type gitClientMock struct {
	CurrentBranchFn        func() (string, error)
	CurrentBranchFnInvoked int
	IsRepoFn               func() bool
	IsRepoFnInvoked        int
	LatestTagFn            func() string
	LatestTagFnInvoked     int
	AncestorTagFn          func(include, exclude string) string
	AncestorTagFnInvoked   int
	SourceBranchFn         func(commitHash string) (string, error)
	SourceBranchFnInvoked  int
}

func initGitClientMock(t *testing.T, latestTag, ancestorTag, currentBranch, sourceBranch, expectedCommitHash string) *gitClientMock {
	return &gitClientMock{
		CurrentBranchFn: func() (string, error) {
			return currentBranch, nil
		},
		IsRepoFn: func() bool {
			return true
		},
		LatestTagFn: func() string {
			return latestTag
		},
		AncestorTagFn: func(include, exclude string) string {
			return ancestorTag
		},
		SourceBranchFn: func(commitHash string) (string, error) {
			assert.Equal(t, expectedCommitHash, commitHash)
			return sourceBranch, nil
		},
	}
}

func (m *gitClientMock) CurrentBranch() (string, error) {
	m.CurrentBranchFnInvoked += 1
	return m.CurrentBranchFn()
}
func (m *gitClientMock) IsRepo() bool {
	m.IsRepoFnInvoked += 1
	return m.IsRepoFn()
}

func (m *gitClientMock) LatestTag() string {
	m.LatestTagFnInvoked += 1
	return m.LatestTagFn()
}

func (m *gitClientMock) AncestorTag(include, exclude string) string {
	m.AncestorTagFnInvoked += 1
	return m.AncestorTagFn(include, exclude)
}

func (m *gitClientMock) SourceBranch(commitHash string) (string, error) {
	m.SourceBranchFnInvoked += 1
	return m.SourceBranchFn(commitHash)
}

func newSemVerPtr(t *testing.T, s string) *semver.Version {
	version, err := semver.New(s)
	require.NoError(t, err)

	return version
}
