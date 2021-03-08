package git

// Client is an empty struct to run git.
type Client struct {
	repoDir string
}

// NewClient creates a new git instance.
func NewClient(repoDir string) *Client {
	return &Client{
		repoDir: repoDir,
	}
}

func (c *Client) CurrentBranch() (string, error) {
	return "", nil
}

func (c *Client) IsRepo() bool {
	return true
}

func (c *Client) LatestTag() (string, error) {
	return "", nil
}

func (c *Client) SourceBranch(commitHash string) (string, error) {
	return "", nil
}
