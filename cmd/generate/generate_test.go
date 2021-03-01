package generate_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/semver-action/cmd/generate"
)

func TestRun_NoPreviousTag(t *testing.T) {
	fp, tearDown := setupTestGitProject(t, 1)
	defer tearDown()

	os.Setenv("GITHUB_SHA", "9c788854a738737498fbd6bec1dc267de6aa7bb6")
	os.Setenv("INPUT_REPO_DIR", fp)
	defer func() {
		os.Unsetenv("GITHUB_SHA")
		os.Unsetenv("INPUT_REPO_DIR")
	}()

	result, err := generate.Run()
	require.NoError(t, err)

	assert.True(t, result.IsPrerelease)
	assert.Equal(t, "v1.0.0-pre.1", result.SemverTag)
	assert.Equal(t, "v0.0.0", result.PreviousTag)
}

func TestRun_SecondMergeIntoDevelop(t *testing.T) {
	fp, tearDown := setupTestGitProject(t, 2)
	defer tearDown()

	os.Setenv("GITHUB_SHA", "bc93de0c53889fd17ae3f6715cae35b40be65830")
	os.Setenv("INPUT_REPO_DIR", fp)
	defer func() {
		os.Unsetenv("GITHUB_SHA")
		os.Unsetenv("INPUT_REPO_DIR")
	}()

	result, err := generate.Run()
	require.NoError(t, err)

	assert.True(t, result.IsPrerelease)
	assert.Equal(t, "v1.1.0-pre.1", result.SemverTag)
	assert.Equal(t, "v1.0.0-pre.1", result.PreviousTag)
}

func TestRun_MergeDevelopIntoMaster(t *testing.T) {
	fp, tearDown := setupTestGitProject(t, 3)
	defer tearDown()

	os.Setenv("GITHUB_SHA", "b2df2617368ba26ba50a78ab736d4875d86a73fa")
	os.Setenv("INPUT_REPO_DIR", fp)
	defer func() {
		os.Unsetenv("GITHUB_SHA")
		os.Unsetenv("INPUT_REPO_DIR")
	}()

	result, err := generate.Run()
	require.NoError(t, err)

	assert.False(t, result.IsPrerelease)
	assert.Equal(t, "v1.1.0", result.SemverTag)
	assert.Equal(t, "v1.1.0-pre.1", result.PreviousTag)
}

func TestRun_MergeHotfixIntoMaster(t *testing.T) {
	fp, tearDown := setupTestGitProject(t, 4)
	defer tearDown()

	os.Setenv("GITHUB_SHA", "58d77591f873a7ba126ecccc2eaf881fe3e467aa")
	os.Setenv("INPUT_REPO_DIR", fp)
	defer func() {
		os.Unsetenv("GITHUB_SHA")
		os.Unsetenv("INPUT_REPO_DIR")
	}()

	result, err := generate.Run()
	require.NoError(t, err)

	assert.False(t, result.IsPrerelease)
	assert.Equal(t, "v1.1.1", result.SemverTag)
	assert.Equal(t, "v1.1.0", result.PreviousTag)
}

func TestRun_BaseVersion(t *testing.T) {
	fp, tearDown := setupTestGitProject(t, 2)
	defer tearDown()

	os.Setenv("GITHUB_SHA", "bc93de0c53889fd17ae3f6715cae35b40be65830")
	os.Setenv("INPUT_BASE_VERSION", "v2.5.17")
	os.Setenv("INPUT_REPO_DIR", fp)
	defer func() {
		os.Unsetenv("GITHUB_SHA")
		os.Unsetenv("INPUT_BASE_VERSION")
		os.Unsetenv("INPUT_REPO_DIR")
	}()

	result, err := generate.Run()
	require.NoError(t, err)

	assert.True(t, result.IsPrerelease)
	assert.Equal(t, "v2.6.0-pre.1", result.SemverTag)
	assert.Equal(t, "v1.0.0-pre.1", result.PreviousTag)
}

func TestRun_BumpMajor(t *testing.T) {
	fp, tearDown := setupTestGitProject(t, 2)
	defer tearDown()

	os.Setenv("GITHUB_SHA", "bc93de0c53889fd17ae3f6715cae35b40be65830")
	os.Setenv("INPUT_BUMP", "major")
	os.Setenv("INPUT_REPO_DIR", fp)
	defer func() {
		os.Unsetenv("GITHUB_SHA")
		os.Unsetenv("INPUT_BUMP")
		os.Unsetenv("INPUT_REPO_DIR")
	}()

	result, err := generate.Run()
	require.NoError(t, err)

	assert.True(t, result.IsPrerelease)
	assert.Equal(t, "v2.0.0-pre.1", result.SemverTag)
	assert.Equal(t, "v1.0.0-pre.1", result.PreviousTag)
}

func TestRun_BumpMinor(t *testing.T) {
	fp, tearDown := setupTestGitProject(t, 2)
	defer tearDown()

	os.Setenv("GITHUB_SHA", "bc93de0c53889fd17ae3f6715cae35b40be65830")
	os.Setenv("INPUT_BUMP", "minor")
	os.Setenv("INPUT_REPO_DIR", fp)
	defer func() {
		os.Unsetenv("GITHUB_SHA")
		os.Unsetenv("INPUT_BUMP")
		os.Unsetenv("INPUT_REPO_DIR")
	}()

	result, err := generate.Run()
	require.NoError(t, err)

	assert.True(t, result.IsPrerelease)
	assert.Equal(t, "v1.1.0-pre.1", result.SemverTag)
	assert.Equal(t, "v1.0.0-pre.1", result.PreviousTag)
}

func TestRun_NoMatchingBranchName(t *testing.T) {
	fp, tearDown := setupTestGitProject(t, 5)
	defer tearDown()

	os.Setenv("GITHUB_SHA", "6369dd5a3c077662cf4abf35ba42a63ddf208d93")
	os.Setenv("INPUT_REPO_DIR", fp)
	defer func() {
		os.Unsetenv("GITHUB_SHA")
		os.Unsetenv("INPUT_REPO_DIR")
	}()

	result, err := generate.Run()
	require.NoError(t, err)

	assert.True(t, result.IsPrerelease)
	assert.Equal(t, "v1.1.0-pre.2", result.SemverTag)
	assert.Equal(t, "v1.1.0-pre.1", result.PreviousTag)
}

func TestRun_BumpPatch(t *testing.T) {
	fp, tearDown := setupTestGitProject(t, 2)
	defer tearDown()

	os.Setenv("GITHUB_SHA", "bc93de0c53889fd17ae3f6715cae35b40be65830")
	os.Setenv("INPUT_BUMP", "patch")
	os.Setenv("INPUT_REPO_DIR", fp)
	defer func() {
		os.Unsetenv("GITHUB_SHA")
		os.Unsetenv("INPUT_BUMP")
		os.Unsetenv("INPUT_REPO_DIR")
	}()

	result, err := generate.Run()
	require.NoError(t, err)

	assert.True(t, result.IsPrerelease)
	assert.Equal(t, "v1.0.1-pre.1", result.SemverTag)
	assert.Equal(t, "v1.0.0-pre.1", result.PreviousTag)
}

func setupTestGitProject(t *testing.T, step int) (fp string, tearDown func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "semver-test")
	require.NoError(t, err)

	copyDir(t, fmt.Sprintf("testdata/git_%v", step), filepath.Join(tmpDir, ".git"))

	return tmpDir, func() { os.RemoveAll(tmpDir) }
}

func copyFile(t *testing.T, source, destination string) {
	input, err := ioutil.ReadFile(source)
	require.NoError(t, err)

	err = ioutil.WriteFile(destination, input, 0600)
	require.NoError(t, err)
}

func copyDir(t *testing.T, src string, dst string) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	require.NoError(t, err)

	if !si.IsDir() {
		return
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}

	if err == nil {
		return
	}

	err = os.MkdirAll(dst, si.Mode())
	require.NoError(t, err)

	entries, err := ioutil.ReadDir(src)
	require.NoError(t, err)

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			copyDir(t, srcPath, dstPath)
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			copyFile(t, srcPath, dstPath)
		}
	}
}
