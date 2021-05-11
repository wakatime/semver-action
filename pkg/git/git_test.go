package git_test

import (
	"errors"
	"testing"

	"github.com/wakatime/semver-action/pkg/git"

	"github.com/alecthomas/assert"
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

	value := gc.LatestTag()

	assert.Equal(t, "v2.4.79", value)
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

	value := gc.LatestTag()

	assert.Empty(t, value)
}

func TestAncestorTag(t *testing.T) {
	tests := map[string]struct {
		IncludePattern string
		ExcludePattern string
		ExpectedTag    string
	}{
		"dev tag only": {
			IncludePattern: "v[0-9]*-dev*",
			ExpectedTag:    "v0.11.1-dev.2",
		},
		"non-dev tag only": {
			IncludePattern: "v[0-9]*",
			ExcludePattern: "v[0-9]*-dev*",
			ExpectedTag:    "v1.2.0",
		},
	}

	gc := git.NewGit("/path/to/repo")
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gc.GitCmd = func(env map[string]string, args ...string) (string, error) {
				assert.Nil(t, env)
				assert.Equal(t, args, []string{"-C", "/path/to/repo", "describe", "--tags", "--abbrev=0", "--match", args[6], "--exclude", args[8]})

				return test.ExpectedTag, nil
			}

			value := gc.AncestorTag(test.IncludePattern, test.ExcludePattern)

			assert.Equal(t, test.ExpectedTag, value)
		})
	}
}

func TestAncestorTag_NoTagFound(t *testing.T) {
	var numCalls int

	gc := git.NewGit("/path/to/repo")
	gc.GitCmd = func(env map[string]string, args ...string) (string, error) {
		numCalls++

		assert.Nil(t, env)

		switch numCalls {
		case 1:
			assert.Equal(t, args, []string{"-C", "/path/to/repo", "describe", "--tags", "--abbrev=0", "--match", args[6], "--exclude", args[8]})
		case 2:
			assert.Equal(t, args, []string{"-C", "/path/to/repo", "rev-list", "--max-parents=0", "HEAD"})
		}

		return "", nil
	}

	value := gc.AncestorTag("", "")

	assert.Empty(t, value)
}
