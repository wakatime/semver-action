package git

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"

	"github.com/apex/log"
)

// Vcs defines the methods to run git.
type Vcs interface {
	Run(args ...string) (string, error)
	Clean(output string, err error) (string, error)
	IsRepo() bool
}

// Git is an empty struct to run git.
type Git struct{}

// NewGit creates a new git instance.
func NewGit() *Git {
	return &Git{}
}

// IsRepo returns true if current folder is a git repository.
func (g *Git) IsRepo() bool {
	out, err := g.Run("rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

// Clean the output.
func (g *Git) Clean(output string, err error) (string, error) {
	output = strings.ReplaceAll(strings.Split(output, "\n")[0], "'", "")
	if err != nil {
		err = errors.New(strings.TrimSuffix(err.Error(), "\n"))
	}
	return output, err
}

// Run runs a git command and returns its output or errors.
func (g *Git) Run(args ...string) (string, error) {
	return runEnv(nil, args...)
}

// runEnv runs a git command with the specified env vars and returns its output or errors.
func runEnv(env map[string]string, args ...string) (string, error) {
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
