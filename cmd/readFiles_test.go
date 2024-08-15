package cmd

import (
	"log/slog"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadDirectoryWithSlash(t *testing.T) {
	assert.Equal(t, "./test_dir", fixTrailingSlashForPath("./test_dir/"), "slash removed at the end")
}

func TestReadDirectoryWithoutSlash(t *testing.T) {
	assert.Equal(t, "./test_dir", fixTrailingSlashForPath("./test_dir"), "no slash at the end")
}

func TestGetNamedMatchesForRegex(t *testing.T) {
	submoduleRegex = regexp.MustCompile("(?P<base_url>.*/.*)//(?P<submodule>.*)")

	//nonSubmoduleMatch := getNamedMatchesForRegex(submoduleRegex, "github.com/hashicorp/example?ref=1.0.0")
	submoduleMatch := getNamedMatchesForRegex(submoduleRegex, "github.com/hashicorp/example//test_1/test_2?ref=1.0.0")
	doubleSlashSubmoduleMatch := getNamedMatchesForRegex(submoduleRegex, "https://github.com/hashicorp/example//test_1/test_2?ref=1.0.0")

	//	assert.NotEmpty(t, nonSubmoduleMatch, "readFiles_test :: getNamedMatchesForRegex :: github.com/hashicorp/example?ref=1.0.0 was not matched for baseurl")
	//	assert.Equal(t, "github.com/hashicorp/example?ref=1.0.0", nonSubmoduleMatch["base_url"], "readFiles_test :: getNamedMatchesForRegex :: unable to extract base url")
	//	assert.Empty(t, nonSubmoduleMatch["submodule"], "readFiles_test :: getNamedMatchesForRegex :: incorrectly extracts submodule")

	assert.NotEmpty(t, submoduleMatch, "readFiles_test :: getNamedMatchesForRegex :: github.com/hashicorp/example//test_1/test_2?ref=1.0.0 was not matched for baseurl")
	assert.Equal(t, "github.com/hashicorp/example", submoduleMatch["base_url"], "readFiles_test :: getNamedMatchesForRegex :: unable to extract base url")
	assert.Equal(t, "test_1/test_2?ref=1.0.0", submoduleMatch["submodule"], "readFiles_test :: getNamedMatchesForRegex :: unable to extract submodule")

	assert.NotEmpty(t, doubleSlashSubmoduleMatch, "readFiles_test :: getNamedMatchesForRegex :: https://github.com/hashicorp/example//test_1/test_2?ref=1.0.0 was not matched for baseurl")
	assert.Equal(t, "https://github.com/hashicorp/example", doubleSlashSubmoduleMatch["base_url"], "readFiles_test :: getNamedMatchesForRegex :: unable to extract base url")
	assert.Equal(t, "test_1/test_2?ref=1.0.0", doubleSlashSubmoduleMatch["submodule"], "readFiles_test :: getNamedMatchesForRegex :: unable to extract submodule")

}

func TestGetTagFromUrl(t *testing.T) {
	assert.Equal(t, "1.0.0", getTagFromUrl("github.com/hashicorp/example?ref=1.0.0"), "ref param not extracted for github.com/hashicorp/example?ref=1.0.0")
	assert.Equal(t, "1.0.0", getTagFromUrl("git@github.com:hashicorp/example?ref=1.0.0"), "ref param not extracted for git@github.com:hashicorp/example?ref=1.0.0")
	assert.Empty(t, getTagFromUrl("github.com/hashicorp/example.git"), "ref param found for github.com/hashicorp/example.git")
	assert.Empty(t, getTagFromUrl("github.com/hashicorp/example?depth=1"), "ref param found")
	assert.Empty(t, getTagFromUrl("https://github.com/hashicorp/example?depth=1"), "ref param found")
	assert.Equal(t, "1.0.1", getTagFromUrl("https://github.com/hashicorp/example?depth=&ref=1.0.1"), "ref param found")

	assert.Equal(t, "2.0.0", getTagFromUrl("https://github.com/hashicorp/example?ref=2.0.0"), "ref param not found for https://github.com/hashicorp/example?ref=2.0.0")
}
func TestExtractRefAndPathRepo(t *testing.T) {
	gitHubExampleRepo, _, _ := extractRefAndPath("github.com/hashicorp/example?ref=1.0.0")
	assert.Equal(t, "github.com/hashicorp/example", gitHubExampleRepo, "readFiles :: extractRefAndPath :: repo :: "+gitHubExampleRepo)

	gitHubSSHExampleRepo, _, _ := extractRefAndPath("git@github.com:hashicorp/example.git")
	assert.Equal(t, "git@github.com:hashicorp/example.git", gitHubSSHExampleRepo, "readFiles :: extractRefAndPath :: repo :: "+gitHubSSHExampleRepo)

	gitHubSSHExampleRepoWithRef, _, _ := extractRefAndPath("git@github.com:hashicorp/example.git?ref=test")
	assert.Equal(t, "git@github.com:hashicorp/example.git", gitHubSSHExampleRepoWithRef, "readFiles :: extractRefAndPath :: repo :: "+gitHubSSHExampleRepo)

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

	//_, gitHubSSHExampleTag, _ := extractRefAndPath("git@github.com:hashicorp/example.git")

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

	gitHubSSHExample, _, gitHubSSHExampleSubmodule := extractRefAndPath("git@github.com:hashicorp/example//module/test")
	assert.Equal(t, "module/test", gitHubSSHExampleSubmodule, "readFiles :: extractRefAndPath :: submodule :: "+gitHubSSHExampleSubmodule)
	assert.Equal(t, "git@github.com:hashicorp/example", gitHubSSHExample, "readFiles :: extractRefAndPath :: repo :: "+gitHubSSHExample)

	gitHubExampleRepo, _, gitHubExampleRepoSubmodule := extractRefAndPath("https://github.com/org/repo//submodules/folder?ref=1.1.1\n")
	assert.Equal(t, "submodules/folder", gitHubExampleRepoSubmodule, "readFiles :: extractRefAndPath :: submodule :: "+gitHubExampleRepoSubmodule)
	assert.Equal(t, "https://github.com/org/repo", gitHubExampleRepo, "readFiles :: extractRefAndPath :: repo :: "+gitHubExampleRepo)

	_, gitHubExampleRepoNoTagSubmodule, gitHubExampleRepoSubmoduleNoTag := extractRefAndPath("https://github.com/org/repo//submodules/folder\n")
	assert.Equal(t, "submodules/folder", gitHubExampleRepoSubmoduleNoTag, "readFiles :: extractRefAndPath :: submodule :: "+gitHubExampleRepoSubmoduleNoTag)
	assert.Equal(t, "", gitHubExampleRepoNoTagSubmodule, "readFiles :: extractRefAndPath :: submodule :: "+gitHubExampleRepoNoTagSubmodule)

}
func TestExtractModuleSource(t *testing.T) {
	nonGitSource := extractModuleSource("source=\"Terraform-VMWare-Modules/vm/vsphere\"")
	assert.Empty(t, nonGitSource, "readFiles :: extractRefAndPath :: repo :: non git terraform source extracted incorrectly")

	gitHubExampleSource := extractModuleSource("source=\"github.com/hashicorp/example?ref=1.0.0\"")
	assert.Equal(t, "github.com/hashicorp/example?ref=1.0.0", gitHubExampleSource, "readFiles :: extractRefAndPath :: repo :: "+gitHubExampleSource)

	gitHubSSHExampleSource := extractModuleSource("source=\"git@github.com:hashicorp/example.git\"")
	assert.Equal(t, "git@github.com:hashicorp/example.git", gitHubSSHExampleSource, "readFiles :: extractRefAndPath :: repo :: "+gitHubSSHExampleSource)

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
