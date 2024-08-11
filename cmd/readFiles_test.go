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

func TestGetTagFromUrl(t *testing.T) {
	assert.Equal(t, "1.0.0", getTagFromUrl("github.com/hashicorp/example?ref=1.0.0"), "ref param not extracted")
	assert.Empty(t, getTagFromUrl("github.com/hashicorp/example.git"), "ref param found")
	assert.Empty(t, getTagFromUrl("github.com/hashicorp/example?depth=1"), "ref param found")
	assert.Empty(t, getTagFromUrl("https://github.com/hashicorp/example?depth=1"), "ref param found")
	assert.Equal(t, "2.0.0", getTagFromUrl("https://github.com/hashicorp/example?ref=2.0.0"), "ref param not found")
}
func TestExtractRefAndPathRepo(t *testing.T) {
	gitHubExampleRepo, _, _ := extractRefAndPath("github.com/hashicorp/example?ref=1.0.0")
	assert.Equal(t, "github.com/hashicorp/example", gitHubExampleRepo, "readFiles :: extractRefAndPath :: repo :: "+gitHubExampleRepo)

	// gitHubSSHExampleRepo, gitHubSSHExampleTag := extractRefAndPath("git@github.com:hashicorp/example.git")
	//assert.Equal(t, "git@github.com:hashicorp/example.git", gitHubSSHExampleRepo, "readFiles :: extractRefAndPath :: repo :: "+gitHubSSHExampleRepo+" :: "+gitHubSSHExampleTag)
	bitbucketExampleRepo, _, _ := extractRefAndPath("bitbucket.org/hashicorp/terraform-consul-aws?ref=1.0.0&test=woho")
	assert.Equal(t, "bitbucket.org/hashicorp/terraform-consul-aws", bitbucketExampleRepo, "readFiles :: extractRefAndPath :: repo :: "+bitbucketExampleRepo)

	genericGitExampleRepo, _, _ := extractRefAndPath("https://example.com/vpc.git?ref=1.1.0&test=woho")
	assert.Equal(t, "https://example.com/vpc.git", genericGitExampleRepo, "readFiles :: extractRefAndPath :: repo :: "+genericGitExampleRepo)

	gitParamExampleRepo, _, _ := extractRefAndPath("https://example.com/vpc.git?depth=1&ref=1.2.0")
	assert.Equal(t, "https://example.com/vpc.git", gitParamExampleRepo, "readFiles :: extractRefAndPath :: repo :: "+gitParamExampleRepo)

	slog.Debug("repos and refs", gitHubExampleRepo, bitbucketExampleRepo, genericGitExampleRepo, gitParamExampleRepo)
}
func TestExtractRefAndPathTag(t *testing.T) {
	_, gitHubExampleTag, _ := extractRefAndPath("github.com/hashicorp/example?ref=1.0.0")
	assert.Equal(t, "1.0.0", gitHubExampleTag, "readFiles :: extractRefAndPath :: tag :: "+gitHubExampleTag)

	_, gitHubExampleNoTag, _ := extractRefAndPath("github.com/hashicorp/example")
	assert.Equal(t, "", gitHubExampleNoTag, "readFiles :: extractRefAndPath :: no tag :: ")

	_, gitHubExampleRepoSubmoduleTag, _ := extractRefAndPath("https://github.com/org/repo//submodules/folder?ref=1.1.1\n")
	assert.Equal(t, "1.1.1", gitHubExampleRepoSubmoduleTag, "readFiles :: extractRefAndPath :: tag :: "+gitHubExampleRepoSubmoduleTag)

	_, gitHubExampleRepoSubmoduleNoTag, _ := extractRefAndPath("https://github.com/org/repo//submodules/folder\n")
	assert.Equal(t, "", gitHubExampleRepoSubmoduleNoTag, "readFiles :: extractRefAndPath :: no tag :: "+gitHubExampleRepoSubmoduleNoTag)

}

func TestExtractRefAndPathSubmodule(t *testing.T) {
	gitHubExample, _, gitHubExampleSubmodule := extractRefAndPath("github.com/hashicorp/example?ref=1.0.0")
	assert.Equal(t, "", gitHubExampleSubmodule, "readFiles :: extractRefAndPath :: submodule :: "+gitHubExampleSubmodule)
	assert.Equal(t, "github.com/hashicorp/example", gitHubExample, "readFiles :: extractRefAndPath :: repo :: "+gitHubExample)

	gitHubExampleRepo, _, gitHubExampleRepoSubmodule := extractRefAndPath("https://github.com/org/repo//submodules/folder?ref=1.1.1\n")
	assert.Equal(t, "submodules/folder", gitHubExampleRepoSubmodule, "readFiles :: extractRefAndPath :: submodule :: "+gitHubExampleRepoSubmodule)
	assert.Equal(t, "https://github.com/org/repo", gitHubExampleRepo, "readFiles :: extractRefAndPath :: repo :: "+gitHubExampleRepo)

	_, gitHubExampleRepoNoTagSubmodule, gitHubExampleRepoSubmoduleNoTag := extractRefAndPath("https://github.com/org/repo//submodules/folder\n")
	assert.Equal(t, "submodules/folder", gitHubExampleRepoSubmoduleNoTag, "readFiles :: extractRefAndPath :: submodule :: "+gitHubExampleRepoSubmoduleNoTag)
	assert.Equal(t, "", gitHubExampleRepoNoTagSubmodule, "readFiles :: extractRefAndPath :: submodule :: "+gitHubExampleRepoNoTagSubmodule)

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

func TestExtractSubmoduleFromSource(t *testing.T) {
	module, noSubmodule := extractSubmoduleFromSource("https://example.com/vpc.git?depth=1&ref=v1.2.0")
	assert.Equal(t, "https://example.com/vpc.git?depth=1&ref=v1.2.0", module, "no submodules and url extract fails")
	assert.Empty(t, noSubmodule, "submodule present")

	module, submoduleWithParams := extractSubmoduleFromSource("https://example.com/testing/vpc//submodule?depth=1&ref=v1.2.0")
	assert.Equal(t, "https://example.com/testing/vpc", module, "url mismatch")
	assert.Equal(t, "submodule?depth=1&ref=v1.2.0", submoduleWithParams)

	module, submoduleWithNoTags := extractSubmoduleFromSource("https://example.com/testing/vpc.git//submodule")
	assert.Equal(t, "https://example.com/testing/vpc.git", module)
	assert.Equal(t, "submodule", submoduleWithNoTags)

	module, submoduleWithDepths := extractSubmoduleFromSource("https://example.com/testing/vpc.git//submodule/folder1/folder2")
	assert.Equal(t, "https://example.com/testing/vpc.git", module)
	assert.Equal(t, "submodule/folder1/folder2", submoduleWithDepths)

}

//gitHubSSHExampleTag, gitHubSSHExampleRepo,
