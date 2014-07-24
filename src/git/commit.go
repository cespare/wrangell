package git

// A Commmit represents a Git commit.
type Commit struct {
	sha *SHA
}

func (c *Commit) SHA() *SHA { return c.sha }
