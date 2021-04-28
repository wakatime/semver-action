package git_test

import (
	"errors"
	"testing"

	"github.com/wakatime/semver-action/pkg/git"

	"github.com/alecthomas/assert"
	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/require"
)

func TestClean(t *testing.T) {
	gc := git.NewGit("/path/to/repo")

	value, err := gc.Clean("'test'", nil)
	require.NoError(t, err)

	assert.Equal(t, "test", value)
}

func TestCleanErr(t *testing.T) {
	gc := git.NewGit("/path/to/repo")

	value, err := gc.Clean("'test'", errors.New("error"))
	require.Error(t, err)

	assert.Equal(t, "test", value)
	assert.EqualError(t, err, "error")
}

func TestCurrentBranch(t *testing.T) {
	gc := git.NewGit("/path/to/repo")
	gc.GitCmd = func(env map[string]string, args ...string) (string, error) {
		assert.Nil(t, env)
		assert.Equal(t, args, []string{"-C", "/path/to/repo", "rev-parse", "--abbrev-ref", "HEAD", "--quiet"})

		return "develop", nil
	}

	value, err := gc.CurrentBranch()
	require.NoError(t, err)

	assert.Equal(t, "develop", value)
}

func TestCurrentBranchErr(t *testing.T) {
	gc := git.NewGit("/path/to/repo")
	gc.GitCmd = func(env map[string]string, args ...string) (string, error) {
		assert.Nil(t, env)
		assert.Equal(t, args, []string{"-C", "/path/to/repo", "rev-parse", "--abbrev-ref", "HEAD", "--quiet"})

		return "", errors.New("error")
	}

	_, err := gc.CurrentBranch()
	require.Error(t, err)

	assert.EqualError(t, err, "could not get current branch: error")
}

func TestSourceBranch(t *testing.T) {
	gc := git.NewGit("/path/to/repo")
	gc.GitCmd = func(env map[string]string, args ...string) (string, error) {
		assert.Nil(t, env)
		assert.Equal(t, args, []string{"-C", "/path/to/repo", "log", "-1", "--pretty=%B", "81918ffc"})

		return "Merge pull request #123 from wakatime/feature/semver-initial", nil
	}

	value, err := gc.SourceBranch("81918ffc")
	require.NoError(t, err)

	assert.Equal(t, "feature/semver-initial", value)
}

func TestSourceBranch_NotValidPullRequestMessage(t *testing.T) {
	gc := git.NewGit("/path/to/repo")
	gc.GitCmd = func(env map[string]string, args ...string) (string, error) {
		assert.Nil(t, env)
		assert.Equal(t, args, []string{"-C", "/path/to/repo", "log", "-1", "--pretty=%B", "81918ffc"})

		return "not valid pull request message", nil
	}

	_, err := gc.SourceBranch("81918ffc")
	require.Error(t, err)

	assert.EqualError(t, err, "no source branch found")
}

func TestSourceBranch_NotValiddBranchName(t *testing.T) {
	gc := git.NewGit("/path/to/repo")
	gc.GitCmd = func(env map[string]string, args ...string) (string, error) {
		assert.Nil(t, env)
		assert.Equal(t, args, []string{"-C", "/path/to/repo", "log", "-1", "--pretty=%B", "81918ffc"})

		return "Merge pull request #123 from semver-initial", nil
	}

	_, err := gc.SourceBranch("81918ffc")
	require.Error(t, err)

	assert.EqualError(t, err, "commit message does not contain expected format: semver-initial")
}

func TestLatestTag(t *testing.T) {
	gc := git.NewGit("/path/to/repo")
	gc.GitCmd = func(env map[string]string, args ...string) (string, error) {
		assert.Nil(t, env)
		assert.Equal(t, args, []string{"-C", "/path/to/repo", "tag", "--points-at", "HEAD", "--sort", "-version:creatordate"})

		return "v2.4.79", nil
	}

	expected, err := semver.New("2.4.79")
	require.NoError(t, err)

	value, err := gc.LatestTag("v")
	require.NoError(t, err)

	assert.Equal(t, expected, value)
}

func TestLatestTag_IncorrectSemver(t *testing.T) {
	gc := git.NewGit("/path/to/repo")
	gc.GitCmd = func(env map[string]string, args ...string) (string, error) {
		assert.Nil(t, env)
		assert.Equal(t, args, []string{"-C", "/path/to/repo", "tag", "--points-at", "HEAD", "--sort", "-version:creatordate"})

		return "v2", nil
	}

	_, err := gc.LatestTag("v")
	require.Error(t, err)

	assert.EqualError(t, err, `failed to parse tag "2" or not valid semantic version: No Major.Minor.Patch elements found`)
}

func TestLatestTag_NoTagFound(t *testing.T) {
	var numCalls int

	gc := git.NewGit("/path/to/repo")
	gc.GitCmd = func(env map[string]string, args ...string) (string, error) {
		numCalls++

		assert.Nil(t, env)

		switch numCalls {
		case 1:
			assert.Equal(t, args, []string{"-C", "/path/to/repo", "tag", "--points-at", "HEAD", "--sort", "-version:creatordate"})
		case 2:
			assert.Equal(t, args, []string{"-C", "/path/to/repo", "describe", "--tags", "--abbrev=0"})
		}

		return "", nil
	}

	value, err := gc.LatestTag("v")
	require.NoError(t, err)

	assert.Nil(t, value)
}

func TestAncestorTag(t *testing.T) {
	gc := git.NewGit("/path/to/repo")
	gc.GitCmd = func(env map[string]string, args ...string) (string, error) {
		assert.Nil(t, env)
		assert.Equal(t, args, []string{"-C", "/path/to/repo", "describe", "--tags", "--abbrev=0", "--match", args[6]})

		return "0.1.3-dev.1", nil
	}

	expected, err := semver.New("0.1.3-dev.1")
	require.NoError(t, err)

	value, err := gc.AncestorTag("v", "v[0-9]*-dev*")
	require.NoError(t, err)

	assert.Equal(t, expected, value)
}

func TestAncestorTag_NoTagFound(t *testing.T) {
	gc := git.NewGit("/path/to/repo")
	gc.GitCmd = func(env map[string]string, args ...string) (string, error) {
		assert.Nil(t, env)
		assert.Equal(t, args, []string{"-C", "/path/to/repo", "describe", "--tags", "--abbrev=0", "--match", args[6]})

		return "", nil
	}

	value, err := gc.AncestorTag("v", "v[0-9]*")
	require.NoError(t, err)

	assert.Nil(t, value)
}

func TestAncestorTag_IncorrectSemver(t *testing.T) {
	gc := git.NewGit("/path/to/repo")
	gc.GitCmd = func(env map[string]string, args ...string) (string, error) {
		assert.Nil(t, env)
		assert.Equal(t, args, []string{"-C", "/path/to/repo", "describe", "--tags", "--abbrev=0", "--match", args[6]})

		return "v2", nil
	}

	_, err := gc.AncestorTag("v", "v[0-9]*-dev*")
	require.Error(t, err)

	assert.EqualError(t, err, `failed to parse tag "2" or not valid semantic version: No Major.Minor.Patch elements found`)
}
