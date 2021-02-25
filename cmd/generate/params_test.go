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
	os.Setenv("INPUT_DRY_RUN", "true")

	params, err := generate.LoadParams()
	require.NoError(t, err)

	assert.Equal(t, "v", params.Prefix)

	os.Unsetenv("INPUT_PREFIX")
	os.Unsetenv("INPUT_DRY_RUN")
}

func TestLoadParams_PrereleaseID(t *testing.T) {
	os.Setenv("INPUT_PRERELEASE_ID", "alpha")
	os.Setenv("INPUT_DRY_RUN", "true")

	params, err := generate.LoadParams()
	require.NoError(t, err)

	assert.Equal(t, "alpha", params.PrereleaseID)

	os.Unsetenv("INPUT_PRERELEASE_ID")
	os.Unsetenv("INPUT_DRY_RUN")
}

func TestLoadParams_MainBranchName(t *testing.T) {
	os.Setenv("INPUT_MAIN_BRANCH_NAME", "master")
	os.Setenv("INPUT_DRY_RUN", "true")

	params, err := generate.LoadParams()
	require.NoError(t, err)

	assert.Equal(t, "master", params.MainBranchName)

	os.Unsetenv("INPUT_MAIN_BRANCH_NAME")
	os.Unsetenv("INPUT_DRY_RUN")
}

func TestLoadParams_DevelopBranchName(t *testing.T) {
	os.Setenv("INPUT_DEVELOP_BRANCH_NAME", "develop")
	os.Setenv("INPUT_DRY_RUN", "true")

	params, err := generate.LoadParams()
	require.NoError(t, err)

	assert.Equal(t, "develop", params.DevelopBranchName)

	os.Unsetenv("INPUT_DEVELOP_BRANCH_NAME")
	os.Unsetenv("INPUT_DRY_RUN")
}

func TestLoadParams_CommitSha(t *testing.T) {
	os.Setenv("GITHUB_SHA", "2f08f7b455ec64741d135216d19d7e0c4dd46458")
	os.Setenv("INPUT_DRY_RUN", "true")

	params, err := generate.LoadParams()
	require.NoError(t, err)

	assert.Equal(t, "2f08f7b455ec64741d135216d19d7e0c4dd46458", params.CommitSha)

	os.Unsetenv("GITHUB_SHA")
	os.Unsetenv("INPUT_DRY_RUN")
}

func TestLoadParams_InvalidCommitSha(t *testing.T) {
	os.Setenv("GITHUB_SHA", "any")

	_, err := generate.LoadParams()
	require.Error(t, err)

	os.Unsetenv("GITHUB_SHA")
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
			os.Setenv("INPUT_DRY_RUN", "true")

			params, err := generate.LoadParams()
			require.NoError(t, err)

			assert.Equal(t, value, params.Bump)

			os.Unsetenv("INPUT_BUMP")
			os.Unsetenv("INPUT_DRY_RUN")
		})
	}
}

func TestLoadParams_AuthToken(t *testing.T) {
	os.Setenv("INPUT_AUTH_TOKEN", "0000000000000000000000000000000000000000")
	os.Setenv("INPUT_DRY_RUN", "true")

	params, err := generate.LoadParams()
	require.NoError(t, err)

	assert.Equal(t, "0000000000000000000000000000000000000000", params.AuthToken)

	os.Unsetenv("INPUT_AUTH_TOKEN")
	os.Unsetenv("INPUT_DRY_RUN")
}

func TestLoadParams_InvalidBump(t *testing.T) {
	os.Setenv("INPUT_BUMP", "invalid")

	_, err := generate.LoadParams()
	require.Error(t, err)

	os.Unsetenv("INPUT_BUMP")
}

func TestLoadParams_Repository(t *testing.T) {
	os.Setenv("GITHUB_REPOSITORY", "wakatime/wakatime-cli")
	os.Setenv("INPUT_DRY_RUN", "true")

	params, err := generate.LoadParams()
	require.NoError(t, err)

	assert.Equal(t, "wakatime", params.Owner)
	assert.Equal(t, "wakatime-cli", params.Repository)

	os.Unsetenv("GITHUB_REPOSITORY")
	os.Unsetenv("INPUT_DRY_RUN")
}

func TestLoadParams_BaseVersion(t *testing.T) {
	os.Setenv("INPUT_BASE_VERSION", "1.2.3")
	os.Setenv("INPUT_AUTH_TOKEN", "0000000000000000000000000000000000000000")
	os.Setenv("INPUT_DRY_RUN", "true")

	params, err := generate.LoadParams()
	require.NoError(t, err)

	var expected = semver.MustParse("1.2.3")

	assert.True(t, expected.EQ(*params.BaseVersion))

	os.Unsetenv("INPUT_BASE_VERSION")
	os.Unsetenv("INPUT_AUTH_TOKEN")
	os.Unsetenv("INPUT_DRY_RUN")
}

func TestLoadParams_InvalidBaseVersion(t *testing.T) {
	os.Setenv("INPUT_BASE_VERSION", "any")

	_, err := generate.LoadParams()
	require.Error(t, err)

	os.Unsetenv("INPUT_BASE_VERSION")
}
