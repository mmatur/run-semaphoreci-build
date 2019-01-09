package types

// Config holds configuration.
type Config struct {
	Owner    string `description:"Repository owner"`
	Project  string `description:"Project"`
	Branch   string `description:"Branch to rebuild"`
	SHA      string `description:"SHA to build"`
	TagEvent bool   `description:"Start the build on tag event only will check $GITHUB_REF env variable"`
}

// NoOption empty struct.
type NoOption struct{}
