package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apex/log"
)

var mergePRRegex = regexp.MustCompile(`Merge pull request #([0-9])+ from (?P<source>.*)+`) // nolint

// Client is an empty struct to run git.
type Client struct {
	repoDir string
	GitCmd  func(env map[string]string, args ...string) (string, error)
}

// NewGit creates a new git instance.
func NewGit(repoDir string) *Client {
	return &Client{
		repoDir: repoDir,
		GitCmd:  gitCmdFn,
	}
}

// gitCmdFn runs a git command with the specified env vars and returns its output or errors.
func gitCmdFn(env map[string]string, args ...string) (string, error) {
	var extraArgs = []string{
		"-c", "log.showSignature=false",
	}

	args = append(extraArgs, args...)
	/* #nosec */
	var cmd = exec.Command("git", args...)

	if env != nil {
		cmd.Env = []string{}
		for k, v := range env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.WithField("args", args).Debug("running git")

	err := cmd.Run()

	log.WithField("stdout", stdout.String()).
		WithField("stderr", stderr.String()).
		Debug("git result")

	if err != nil {
		return "", errors.New(stderr.String())
	}

	return stdout.String(), nil
}

// Clean the output.
func (Client) Clean(output string, err error) (string, error) {
	output = strings.ReplaceAll(strings.Split(output, "\n")[0], "'", "")

	if err != nil {
		err = errors.New(strings.TrimSuffix(err.Error(), "\n"))
	}

	return output, err
}

// Run runs a git command and returns its output or errors.
func (c *Client) Run(args ...string) (string, error) {
	return c.GitCmd(nil, args...)
}

// MakeSafe adds safe.directory global config.
func (c *Client) MakeSafe() error {
	dir, err := filepath.Abs(c.repoDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for: %s", c.repoDir)
	}

	_, err = c.Run("config", "--global", "--add", "safe.directory", dir)
	if err != nil {
		return fmt.Errorf("failed to set safe current directory")
	}

	return nil
}

// IsRepo returns true if current folder is a git repository.
func (c *Client) IsRepo() bool {
	out, err := c.Run("rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

// CurrentBranch returns the current branch checked out.
func (c *Client) CurrentBranch() (string, error) {
	dest, err := c.Clean(c.Run("-C", c.repoDir, "rev-parse", "--abbrev-ref", "HEAD", "--quiet"))
	if err != nil {
		return "", fmt.Errorf("could not get current branch: %s", err)
	}

	return dest, nil
}

// SourceBranch tries to get branch from commit message.
func (c *Client) SourceBranch(commitHash string) (string, error) {
	message, err := c.Clean(c.Run("-C", c.repoDir, "log", "-1", "--pretty=%B", commitHash))
	if err != nil {
		return "", fmt.Errorf("could not get message from commit: %s", err)
	}

	match := mergePRRegex.FindStringSubmatch(message)

	paramsMap := make(map[string]string)

	for i, name := range mergePRRegex.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}

	if len(paramsMap) == 0 || paramsMap["source"] == "" {
		return "", errors.New("no source branch found")
	}

	splitted := strings.SplitN(paramsMap["source"], "/", 2)

	if len(splitted) < 2 {
		return "", fmt.Errorf("commit message does not contain expected format: %s", paramsMap["source"])
	}

	return splitted[1], nil
}

// LatestTag returns the latest tag if found.
func (c *Client) LatestTag() string {
	var result string

	commitSha, _ := c.Clean(c.Run("-C", c.repoDir, "rev-list", "--tags", "--max-count=1"))
	if commitSha != "" {
		result, _ = c.Clean(c.Run("-C", c.repoDir, "describe", "--tags", commitSha))
	}

	return result
}

// AncestorTag returns the previous tag that matches specific pattern if found.
func (c *Client) AncestorTag(include, exclude, branch string) string {
	result, _ := c.Clean(c.Run(
		"-C", c.repoDir, "describe", "--tags", "--abbrev=0",
		"--match", include, "--exclude", exclude, branch))
	if result == "" {
		result, _ = c.Clean(c.Run("-C", c.repoDir, "rev-list", "--max-parents=0", "HEAD"))
	}

	return result
}
