package cmd

import (
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestReadDirectoryWithSlash(t *testing.T) {
	assert.Equal(t, "./test_dir", fixTrailingSlashForPath("./test_dir/"), "slash removed at the end")
}

func TestReadDirectoryWithoutSlash(t *testing.T) {
	assert.Equal(t, "./test_dir", fixTrailingSlashForPath("./test_dir"), "no slash at the end")
}

func TestExtractRefAndPath(t *testing.T) {
	gitHubExampleRepo, gitHubExampleTag := extractRefAndPath("github.com/hashicorp/example?ref=1.0.0")
	assert.Equal(t, "github.com/hashicorp/example", gitHubExampleRepo, "readFiles :: extractRefAndPath :: repo :: "+gitHubExampleRepo)
	assert.Equal(t, "1.0.0", gitHubExampleTag, "readFiles :: extractRefAndPath :: tag :: "+gitHubExampleTag)

	gitHubExampleRepoNoTag, gitHubExampleNoTag := extractRefAndPath("github.com/hashicorp/example")
	assert.Equal(t, "github.com/hashicorp/example", gitHubExampleRepoNoTag, "readFiles :: extractRefAndPath :: repo :: "+gitHubExampleRepoNoTag)
	assert.Empty(t, gitHubExampleNoTag, "readFiles :: extractRefAndPath :: tag :: "+gitHubExampleNoTag)

	// gitHubSSHExampleRepo, gitHubSSHExampleTag := extractRefAndPath("git@github.com:hashicorp/example.git")
	//assert.Equal(t, "git@github.com:hashicorp/example.git", gitHubSSHExampleRepo, "readFiles :: extractRefAndPath :: repo :: "+gitHubSSHExampleRepo+" :: "+gitHubSSHExampleTag)
	bitbucketExampleRepo, bitbucketExampleTag := extractRefAndPath("bitbucket.org/hashicorp/terraform-consul-aws?ref=1.0.0&test=woho")
	assert.Equal(t, "bitbucket.org/hashicorp/terraform-consul-aws", bitbucketExampleRepo, "readFiles :: extractRefAndPath :: repo :: "+bitbucketExampleRepo)
	assert.Equal(t, "1.0.0", bitbucketExampleTag, "readFiles :: extractRefAndPath :: tag :: "+bitbucketExampleTag)
	genericGitExampleRepo, genericGitExampleTag := extractRefAndPath("https://example.com/vpc.git?ref=1.1.0&test=woho")
	assert.Equal(t, "https://example.com/vpc.git", genericGitExampleRepo, "readFiles :: extractRefAndPath :: repo :: "+genericGitExampleRepo)
	assert.Equal(t, "1.1.0", genericGitExampleTag, "readFiles :: extractRefAndPath :: tag :: "+genericGitExampleTag)

	gitParamExampleRepo, gitParamExampleTag := extractRefAndPath("https://example.com/vpc.git?depth=1&ref=1.2.0")
	assert.Equal(t, "https://example.com/vpc.git", gitParamExampleRepo, "readFiles :: extractRefAndPath :: repo :: "+gitParamExampleRepo)
	assert.Equal(t, "1.2.0", gitParamExampleTag, "readFiles :: extractRefAndPath :: tag :: "+gitParamExampleTag)

	slog.Debug("repos and refs", gitHubExampleTag, gitHubExampleRepo, bitbucketExampleTag, bitbucketExampleRepo, genericGitExampleTag, genericGitExampleRepo, gitParamExampleTag, gitParamExampleRepo)
}

func TestExtractModuleSource(t *testing.T) {
	gitHubExampleSource := extractModuleSource("source=\"github.com/hashicorp/example?ref=1.0.0\"")
	assert.Equal(t, "github.com/hashicorp/example?ref=1.0.0", gitHubExampleSource, "readFiles :: extractRefAndPath :: repo :: "+gitHubExampleSource)
	// gitHubSSHExampleSource := extractModuleSource("source=\"git@github.com:hashicorp/example.git")
	//assert.Equal(t, "git@github.com:hashicorp/example.git", gitHubSSHExampleSource, "readFiles :: extractRefAndPath :: repo :: "+gitHubSSHExampleSource+" :: "+gitHubSSHExampleTag)

	gitGitHubExampleSource := extractModuleSource("source=\"git::https://github.com/test_repo_labala?ref=1.3.1\"")
	assert.Equal(t, "https://github.com/test_repo_labala?ref=1.3.1", gitGitHubExampleSource, "readFiles :: extractRefAndPath :: repo :: "+gitGitHubExampleSource)

	bitbucketExampleSource := extractModuleSource("source=\"bitbucket.org/hashicorp/terraform-consul-aws?ref=1.0.0&test=woho\"")
	assert.Equal(t, "bitbucket.org/hashicorp/terraform-consul-aws?ref=1.0.0&test=woho", bitbucketExampleSource, "readFiles :: extractRefAndPath :: repo :: "+bitbucketExampleSource)
	genericGitExampleSource := extractModuleSource("source=\"https://example.com/vpc.git?ref=1.1.0&test=woho\"")
	assert.Equal(t, "https://example.com/vpc.git?ref=1.1.0&test=woho", genericGitExampleSource, "readFiles :: extractRefAndPath :: repo :: "+genericGitExampleSource)
	gitParamExampleSource := extractModuleSource("source=\"https://example.com/vpc.git?depth=1&ref=v1.2.0\"")
	assert.Equal(t, "https://example.com/vpc.git?depth=1&ref=v1.2.0", gitParamExampleSource, "readFiles :: extractRefAndPath :: repo :: "+gitParamExampleSource)
	slog.Debug("repos and refs", gitHubExampleSource, bitbucketExampleSource, genericGitExampleSource, gitParamExampleSource)
}

//gitHubSSHExampleTag, gitHubSSHExampleRepo,
