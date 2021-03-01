package generate_test

import (
	"os"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/wakatime/semver-action/cmd/generate"

	"github.com/alecthomas/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadParams_Prefix(t *testing.T) {
	os.Setenv("INPUT_PREFIX", "v")
	defer os.Unsetenv("INPUT_PREFIX")

	params, err := generate.LoadParams()
	require.NoError(t, err)

	assert.Equal(t, "v", params.Prefix)
}

func TestLoadParams_PrereleaseID(t *testing.T) {
	os.Setenv("INPUT_PRERELEASE_ID", "alpha")
	defer os.Unsetenv("INPUT_PRERELEASE_ID")

	params, err := generate.LoadParams()
	require.NoError(t, err)

	assert.Equal(t, "alpha", params.PrereleaseID)
}

func TestLoadParams_MainBranchName(t *testing.T) {
	os.Setenv("INPUT_MAIN_BRANCH_NAME", "master")
	defer os.Unsetenv("INPUT_MAIN_BRANCH_NAME")

	params, err := generate.LoadParams()
	require.NoError(t, err)

	assert.Equal(t, "master", params.MainBranchName)
}

func TestLoadParams_DevelopBranchName(t *testing.T) {
	os.Setenv("INPUT_DEVELOP_BRANCH_NAME", "develop")
	defer os.Unsetenv("INPUT_DEVELOP_BRANCH_NAME")

	params, err := generate.LoadParams()
	require.NoError(t, err)

	assert.Equal(t, "develop", params.DevelopBranchName)
}

func TestLoadParams_CommitSha(t *testing.T) {
	os.Setenv("GITHUB_SHA", "2f08f7b455ec64741d135216d19d7e0c4dd46458")
	defer os.Unsetenv("GITHUB_SHA")

	params, err := generate.LoadParams()
	require.NoError(t, err)

	assert.Equal(t, "2f08f7b455ec64741d135216d19d7e0c4dd46458", params.CommitSha)
}

func TestLoadParams_RepoDir(t *testing.T) {
	os.Setenv("INPUT_REPO_DIR", "/var/tmp/wakatime-cli")
	defer os.Unsetenv("INPUT_REPO_DIR")

	params, err := generate.LoadParams()
	require.NoError(t, err)

	assert.Equal(t, "/var/tmp/wakatime-cli", params.RepoDir)
}

func TestLoadParams_InvalidCommitSha(t *testing.T) {
	os.Setenv("GITHUB_SHA", "any")
	defer os.Unsetenv("GITHUB_SHA")

	_, err := generate.LoadParams()
	require.Error(t, err)
}

func TestLoadParams_Bump(t *testing.T) {
	tests := map[string]string{
		"auto":  "auto",
		"major": "major",
		"minor": "minor",
		"patch": "patch",
		"empty": "auto",
	}

	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv("INPUT_BUMP", value)
			defer os.Unsetenv("INPUT_BUMP")

			params, err := generate.LoadParams()
			require.NoError(t, err)

			assert.Equal(t, value, params.Bump)
		})
	}
}

func TestLoadParams_InvalidBump(t *testing.T) {
	os.Setenv("INPUT_BUMP", "invalid")
	defer os.Unsetenv("INPUT_BUMP")

	_, err := generate.LoadParams()
	require.Error(t, err)
}

func TestLoadParams_BaseVersion(t *testing.T) {
	os.Setenv("INPUT_BASE_VERSION", "1.2.3")
	defer os.Unsetenv("INPUT_BASE_VERSION")

	params, err := generate.LoadParams()
	require.NoError(t, err)

	var expected = semver.MustParse("1.2.3")

	assert.True(t, expected.EQ(*params.BaseVersion))
}

func TestLoadParams_InvalidBaseVersion(t *testing.T) {
	os.Setenv("INPUT_BASE_VERSION", "any")
	defer os.Unsetenv("INPUT_BASE_VERSION")

	_, err := generate.LoadParams()
	require.Error(t, err)
}
