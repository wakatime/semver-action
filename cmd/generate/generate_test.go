package generate_test

import (
	"github.com/wakatime/semver-action/cmd/generate"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateTag(t *testing.T) {
	tests := map[string]struct {
		CurrentBranch string
		LatestTag     string
		SourceBranch  string
		Params        generate.Params
		Result        generate.Result
	}{
		"develop feature increment": {
			CurrentBranch: "develop",
			LatestTag:     "v0.2.1",
			SourceBranch:  "feature/blazor-language-constant",
			Params: generate.Params{
				CommitSha:         "81918ffc",
				RepoDir:           "github.com/wakatime/semver-action",
				Bump:              "auto",
				BaseVersion:       &semver.Version{},
				Prefix:            "v",
				PrereleaseID:      "alpha",
				MainBranchName:    "master",
				DevelopBranchName: "develop",
			},
			Result: generate.Result{
				PreviousTag:  "v0.2.1",
				SemverTag:    "v0.3.0-alpha.1",
				IsPrerelease: false,
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

func initGitClientMock(t *testing.T, latestTag, currentBranch, sourceBranch, expectedCommitHash string) *gitClientMock {
	return &gitClientMock{
		CurrentBranchFn: func() (string, error) {
			return currentBranch, nil
		},
		IsRepoFn: func() bool {
			return true
		},
		LatestTagFn: func() (string, error) {
			return latestTag, nil
		},
		SourceBranchFn: func(commitHash string) (string, error) {
			assert.Equal(t, expectedCommitHash, commitHash)
			return sourceBranch, nil
		},
	}
}

type gitClientMock struct {
	CurrentBranchFn        func() (string, error)
	CurrentBranchFnInvoked int
	IsRepoFn               func() bool
	IsRepoFnInvoked        int
	LatestTagFn            func() (string, error)
	LatestTagFnInvoked     int
	SourceBranchFn         func(commitHash string) (string, error)
	SourceBranchFnInvoked  int
}

func (m *gitClientMock) CurrentBranch() (string, error) {
	m.CurrentBranchFnInvoked += 1
	return m.CurrentBranchFn()
}

func (m *gitClientMock) IsRepo() bool {
	m.IsRepoFnInvoked += 1
	return m.IsRepoFn()
}

func (m *gitClientMock) LatestTag() (string, error) {
	m.LatestTagFnInvoked += 1
	return m.LatestTagFn()
}

func (m *gitClientMock) SourceBranch(commitHash string) (string, error) {
	m.SourceBranchFnInvoked += 1
	return m.SourceBranchFn(commitHash)
}
