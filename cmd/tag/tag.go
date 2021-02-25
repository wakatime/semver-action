package tag

import (
	"context"
	"fmt"

	"github.com/apex/log"
	"github.com/google/go-github/v33/github"
)

// Params contains semver tag parameters.
type Params struct {
	CommitSha  string
	Client     *github.Client
	Owner      string
	Repository string
	TagMessage string
	Tag        string
}

// Create creates a new tag in the repository.
func Create(p Params) error {
	objType := "commit"
	ctx := context.Background()

	createdTag, resp, err := p.Client.Git.CreateTag(ctx, p.Owner, p.Repository, &github.Tag{
		Tag:     &p.Tag,
		Message: &p.TagMessage,
		SHA:     &p.CommitSha,
		Object: &github.GitObject{
			Type: &objType,
			SHA:  &p.CommitSha,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create tag %q: %s", p.Tag, err)
	}

	if resp.StatusCode != 201 {
		return fmt.Errorf("failed to create tag %q: with status code: %d", p.Tag, resp.StatusCode)

	}

	log.Debugf("created tag %s", p.Tag)

	refs := "refs/tags/" + p.Tag

	_, resp, err = p.Client.Git.CreateRef(ctx, p.Owner, p.Repository, &github.Reference{
		Ref:    &refs,
		Object: createdTag.Object,
	})

	if err != nil {
		return fmt.Errorf("failed to push tag %q: %s", p.Tag, err)
	}

	if resp.StatusCode != 201 {
		return fmt.Errorf("failed to push tag %q: with status code: %d", p.Tag, resp.StatusCode)
	}

	log.Debugf("pushed tag %s", p.Tag)

	return nil
}
