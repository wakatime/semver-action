package generate

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert"
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

	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			method, version := determineBumpStrategy(value.Bump, value.SourceBranch, value.DestBranch, "master", "develop")

			assert.Equal(t, value.ExpectedMethod, method)
			assert.Equal(t, value.ExpectedVersion, version)
		})
	}
}

func TestGetLatestTagOrDefault_ReturnLatest(t *testing.T) {
	g := &vcsMock{
		RunFn: func(args ...string) (string, error) {
			return "v1.2.3", nil
		},
		CleanFn: func(output string, err error) (string, error) {
			return output, err
		},
	}

	tag, err := getLatestTagOrDefault(g, "v", "")
	require.NoError(t, err)

	assert.Equal(t, "1.2.3", tag.String())
}

func TestGetLatestTagOrDefault_ReturnDefault(t *testing.T) {
	g := &vcsMock{
		RunFn: func(args ...string) (string, error) {
			return "", nil
		},
		CleanFn: func(output string, err error) (string, error) {
			return output, err
		},
	}

	tag, err := getLatestTagOrDefault(g, "v", "")
	require.NoError(t, err)

	assert.Equal(t, "0.0.0", tag.String())
}

func TestGetDestBranchFromCommit(t *testing.T) {
	g := &vcsMock{
		RunFn: func(args ...string) (string, error) {
			return "develop", nil
		},
		CleanFn: func(output string, err error) (string, error) {
			return output, err
		},
	}

	dest, err := getDestBranchFromCommit(g, "")
	require.NoError(t, err)

	assert.Equal(t, "develop", dest)
}

func TestGetDestBranchFromCommitErr(t *testing.T) {
	g := &vcsMock{
		RunFn: func(args ...string) (string, error) {
			return "", fmt.Errorf("error")
		},
		CleanFn: func(output string, err error) (string, error) {
			return output, err
		},
	}

	_, err := getDestBranchFromCommit(g, "")

	assert.Error(t, err)
	assert.Equal(t, "could not get current branch: error", err.Error())
}

func TestGetSourceBranchFromCommit(t *testing.T) {
	g := &vcsMock{
		RunFn: func(args ...string) (string, error) {
			return "Merge pull request #2 from wakatime/feature/authentcation", nil
		},
		CleanFn: func(output string, err error) (string, error) {
			return output, err
		},
	}

	source, err := getSourceBranchFromCommit(g, "", "")
	require.NoError(t, err)

	assert.Equal(t, "feature/authentcation", source)
}

func TestGetSourceBranchFromCommitErr(t *testing.T) {
	tests := map[string]struct {
		Message      string
		VcsMock      *vcsMock
		ErrorMessage string
	}{
		"error running git": {
			VcsMock: &vcsMock{
				RunFn: func(args ...string) (string, error) {
					return "", fmt.Errorf("error")
				},
				CleanFn: func(output string, err error) (string, error) {
					return output, err
				},
			},
			ErrorMessage: "could not get message from commit: error",
		},
		"no source branch": {
			VcsMock: &vcsMock{
				RunFn: func(args ...string) (string, error) {
					return "Merge pull request #5 from ", nil
				},
				CleanFn: func(output string, err error) (string, error) {
					return output, err
				},
			},
			ErrorMessage: "no source branch found",
		},
		"non valid message format": {
			VcsMock: &vcsMock{
				RunFn: func(args ...string) (string, error) {
					return "Merge pull request #5 from feature-authentication", nil
				},
				CleanFn: func(output string, err error) (string, error) {
					return output, err
				},
			},
			ErrorMessage: "commit message does not contain expected format: feature-authentication",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := getSourceBranchFromCommit(test.VcsMock, "", "")

			assert.Error(t, err)
			assert.Equal(t, test.ErrorMessage, err.Error())
		})
	}
}

type vcsMock struct {
	RunFn    func(args ...string) (string, error)
	CleanFn  func(output string, err error) (string, error)
	IsRepoFn func() bool
}

func (g *vcsMock) Run(args ...string) (string, error) {
	return g.RunFn(args...)
}

func (g *vcsMock) Clean(output string, err error) (string, error) {
	return g.CleanFn(output, err)
}

func (g *vcsMock) IsRepo() bool {
	return g.IsRepoFn()
}
