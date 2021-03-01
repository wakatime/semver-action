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
		SourceBranch  string
		Params        generate.Params
		Result        generate.Result
	}{
		"no previous tag": {
			CurrentBranch: "develop",
			LatestTag:     "",
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
				SemverTag:    "v1.0.0-alpha.1",
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
	LatestTagFn            func(prefix string) (*semver.Version, error)
	LatestTagFnInvoked     int
	SourceBranchFn         func(commitHash string) (string, error)
	SourceBranchFnInvoked  int
}

func initGitClientMock(t *testing.T, latestTag, currentBranch, sourceBranch, expectedCommitHash string) *gitClientMock {
	return &gitClientMock{
		CurrentBranchFn: func() (string, error) {
			return currentBranch, nil
		},
		IsRepoFn: func() bool {
			return true
		},
		LatestTagFn: func(prefix string) (*semver.Version, error) {
			if latestTag == "" {
				return nil, nil
			}

			version, err := semver.New(latestTag)
			require.NoError(t, err)

			return version, nil
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

func (m *gitClientMock) LatestTag(prefix string) (*semver.Version, error) {
	m.LatestTagFnInvoked += 1
	return m.LatestTagFn(prefix)
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
