package repos

// Repo represents a git repository.
type Repo struct {
	// Path is the file path on the local machine to the git repository.
	Path string
	// URL is the location of the remote git repository.
	URL string
}
