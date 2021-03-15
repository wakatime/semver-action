package generate

import (
	"testing"

	"github.com/alecthomas/assert"
	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/require"
)

func TestDetermineBumpStrategy(t *testing.T) {
	tests := map[string]struct {
		SourceBranch    string
		DestBranch      string
		Bump            string
		ExpectedMethod  string
		ExpectedVersion string
	}{
		"source branch bugfix, dest branch develop and auto bump": {
			SourceBranch:    "bugfix/some",
			DestBranch:      "develop",
			Bump:            "auto",
			ExpectedMethod:  "build",
			ExpectedVersion: "patch",
		},
		"source branch feature, dest branch develop and auto bump": {
			SourceBranch:    "feature/some",
			DestBranch:      "develop",
			Bump:            "auto",
			ExpectedMethod:  "build",
			ExpectedVersion: "minor",
		},
		"source branch major, dest branch develop and auto bump": {
			SourceBranch:    "major/some",
			DestBranch:      "develop",
			Bump:            "auto",
			ExpectedMethod:  "build",
			ExpectedVersion: "major",
		},
		"source branch hotfix, dest branch master and auto bump": {
			SourceBranch:   "hotfix/some",
			DestBranch:     "master",
			Bump:           "auto",
			ExpectedMethod: "hotfix",
		},
		"source branch develop, dest branch master and auto bump": {
			SourceBranch:   "develop",
			DestBranch:     "master",
			Bump:           "auto",
			ExpectedMethod: "final",
		},
		"not a valid source branch prefix and auto bump": {
			SourceBranch:   "some-branch",
			Bump:           "auto",
			ExpectedMethod: "build",
		},
		"patch bump": {
			Bump:           "patch",
			ExpectedMethod: "patch",
		},
		"minor bump": {
			Bump:           "minor",
			ExpectedMethod: "minor",
		},
		"major bump": {
			Bump:           "major",
			ExpectedMethod: "major",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			method, version := determineBumpStrategy(test.Bump, test.SourceBranch, test.DestBranch, "master", "develop")

			assert.Equal(t, test.ExpectedMethod, method)
			assert.Equal(t, test.ExpectedVersion, version)
		})
	}
}

func TestDetermineLatestTag(t *testing.T) {
	tests := map[string]struct {
		LatestTag string
		Prefix    string
		Expected  string
	}{
		"valid latest tag": {
			LatestTag: "v1.2.4",
			Prefix:    "v",
			Expected:  "1.2.4",
		},
		"default tag": {
			LatestTag: "",
			Prefix:    "v",
			Expected:  "0.0.0",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gc := initGitClientMock(
				t,
				test.LatestTag,
			)

			result, err := determineLatestTag(test.Prefix, gc)
			require.NoError(t, err)

			expected, err := semver.New(test.Expected)
			require.NoError(t, err)

			assert.Equal(t, expected, result)
		})
	}
}

func TestDetermineLatestTagErr(t *testing.T) {
	gc := initGitClientMock(
		t,
		"v2",
	)

	_, err := determineLatestTag("v", gc)

	assert.EqualError(t, err, `failed to parse tag "2" or not valid semantic version: No Major.Minor.Patch elements found`)
}

type gitClientMock struct {
	CurrentBranchFn        func() (string, error)
	CurrentBranchFnInvoked int
	IsRepoFn               func() bool
	IsRepoFnInvoked        int
	LatestTagFn            func() string
	LatestTagFnInvoked     int
	SourceBranchFn         func(commitHash string) (string, error)
	SourceBranchFnInvoked  int
}

func initGitClientMock(t *testing.T, latestTag string) *gitClientMock {
	return &gitClientMock{
		LatestTagFn: func() string {
			return latestTag
		},
	}
}

func (m *gitClientMock) CurrentBranch() (string, error) {
	m.CurrentBranchFnInvoked++
	return m.CurrentBranchFn()
}
func (m *gitClientMock) IsRepo() bool {
	m.IsRepoFnInvoked++
	return m.IsRepoFn()
}

func (m *gitClientMock) SourceBranch(commitHash string) (string, error) {
	m.SourceBranchFnInvoked++
	return m.SourceBranchFn(commitHash)
}

func (m *gitClientMock) LatestTag() string {
	m.LatestTagFnInvoked++
	return m.LatestTagFn()
}
