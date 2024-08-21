package cmd

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/thundersparkf/samwise/cmd/errorHandlers"
)

/* func TestGitAuthenticationGenerator(t *testing.T) {
	// TODO: Add tests to check key present and getting ssh type for urls with @
	// TODO: Add tests for key not present panic
	// TODO: Add tests for @ not present in url and username and password retrieved
	// TODO: Add tests for user and password not passed so they return empty strings
	// TODO: Incorrect format private key parsing panic
	//viper.Set("git_ssh_key_path", "/Users/agastya/.ssh/github-thundersparkf")
	//viper.Set("git_key", "test")
	//viper.Set("git_user", "test")
	sshAuth := gitAuthGenerator("git@github.com:Darth-Tech/stack.git")
	fmt.Println(sshAuth.Name())
	// sshAuthProtocol := gitAuthGenerator("ssh://git@github.com:Darth-Tech/stack.git")
	// httpsGithubAuth := gitAuthGenerator("https://github.com/Darth-Tech/terraform-modules")
	// httpsGithubAuthTag := gitAuthGenerator("https://github.com/Darth-Tech/terraform-modules?ref=v0.1.0")
	// bitbucketAuth := gitAuthGenerator("https://bitbucket.org.com/Darth-Tech/terraform-modules")
	// randomDomain := gitAuthGenerator("https://example.com/Darth-Tech/terraform-modules?ref=v0.1.0")

	// assert.Equal(t, "test", sshAuth.Name())
	// malformed := gitAuthGenerator("github.com:Darth-Tech/stack.git")
	// empty := gitAuthGenerator("")

} */

func TestParseGitUrl(t *testing.T) {
	sshAuthProtocol := parseGitUrl("ssh://git@github.com:Darth-Tech/stack.git")
	assert.NotEmpty(t, sshAuthProtocol, "ssh auth protocol parsing empty")
	assert.Equal(t, "git@github.com:Darth-Tech/stack.git", sshAuthProtocol, "ssh auth protocol parsing")

	sshNoProtocolAuthProtocol := parseGitUrl("git@github.com:Darth-Tech/stack.git")
	assert.NotEmpty(t, sshNoProtocolAuthProtocol, "ssh auth without protocol parsing empty")
	assert.Equal(t, "git@github.com:Darth-Tech/stack.git", sshNoProtocolAuthProtocol, "ssh auth without protocol parsing not matching")

	httpsGithubAuth := parseGitUrl("https://github.com/Darth-Tech/terraform-modules")
	assert.NotEmpty(t, httpsGithubAuth, "basic auth with https protocol empty url")
	assert.Equal(t, "https://github.com/Darth-Tech/terraform-modules", httpsGithubAuth, "basic auth with https protocol not matching")

	httpsGithubAuthTag := parseGitUrl("github.com/Darth-Tech/terraform-modules")
	assert.NotEmpty(t, httpsGithubAuthTag, "basic auth without protocol empty url")
	assert.Equal(t, "https://github.com/Darth-Tech/terraform-modules", httpsGithubAuthTag, "basic auth without protocol url not matching")

	bitbucketAuth := parseGitUrl("https://bitbucket.org.com/Darth-Tech/terraform-modules")
	assert.NotEmpty(t, bitbucketAuth, "basic auth with https bitbucket protocol empty url")
	assert.Equal(t, "https://bitbucket.org.com/Darth-Tech/terraform-modules", bitbucketAuth, "basic auth with https bitbucket protocol not matching")

	randomDomain := parseGitUrl("https://example.com/Darth-Tech/terraform-modules")
	assert.NotEmpty(t, randomDomain, "basic auth with https random domain protocol empty url")
	assert.Equal(t, "https://example.com/Darth-Tech/terraform-modules", randomDomain, "basic auth with https random domain protocol not matching")

	httpsGithubAuthGitProtovcol := parseGitUrl("git::https://github.com/Darth-Tech/terraform-modules")
	assert.NotEmpty(t, httpsGithubAuthGitProtovcol, "basic auth with https and git:: random domain protocol empty url")
	assert.Equal(t, "https://github.com/Darth-Tech/terraform-modules", httpsGithubAuthGitProtovcol, "basic auth with https and git:: random domain protocol not matching")
}

// TODO: Add tests for ssh cloning
func TestHappyClonePublicRepo(t *testing.T) {
	r, err := cloneRepo("https://github.com/thundersparkf/adlnet-lrs-py3.git")
	assert.NotEmpty(t, r, "readGit_files.go :: cloneRepo :: repository is nil")
	assert.Empty(t, err, "readGit_files.go :: cloneRepo :: err is not nil :: ")
	head, _ := r.Head()
	assert.Equal(t, "refs/heads/master", head.Name().String(), "readGit_files.go :: cloneRepo :: repository's head is ref/heads/master")

}

func TestHappySSHClonePublicRepo(t *testing.T) {
	viper.Set("git_ssh_key_path", os.Getenv("SAMWISE_CLI_GIT_SSH_KEY_PATH"))
	r, err := cloneRepo("git@github.com:Darth-Tech/stack.git")
	assert.NotEmpty(t, r, "readGit_files.go :: cloneRepo :: repository is nil")
	assert.Empty(t, err, "readGit_files.go :: cloneRepo :: err is not nil :: ")

	r1, err := cloneRepo("ssh://git@github.com:Darth-Tech/stack.git")
	assert.NotEmpty(t, r1, "readGit_files.go :: cloneRepo :: repository is nil")
	assert.Empty(t, err, "readGit_files.go :: cloneRepo :: err is not nil :: ")

	head, err := r.Head()
	assert.Empty(t, err, "readGit_files.go :: cloneRepo :: error not empty getting head of branch :: error :: ")
	assert.Equal(t, "refs/heads/master", head.Name().String(), "readGit_files.go :: cloneRepo :: repository's head is ref/heads/master")

	head1, err := r1.Head()
	assert.Empty(t, err, "readGit_files.go :: cloneRepo :: error not empty getting head of branch :: error :: ")
	assert.Equal(t, "refs/heads/master", head1.Name().String(), "readGit_files.go :: cloneRepo :: repository's head is ref/heads/master")

}

func TestUnhappyNoProtocolClonePublicRepo(t *testing.T) {
	r, err := cloneRepo("github.com/thundersparkf/adlnet-lrs-py3.git")
	assert.NotEmpty(t, r, "readGit_files.go :: cloneRepo :: repository is nil")
	assert.Empty(t, err, "readGit_files.go :: cloneRepo :: err is not nil :: ")
	head, err := r.Head()
	assert.Empty(t, err, "readGit_files.go :: cloneRepo :: err is not nil :: ")
	assert.Equal(t, "refs/heads/master", head.Name().String(), "readGit_files.go :: cloneRepo :: repository's head is ref/heads/master")

}

func TestUnhappyClonePublicRepo(t *testing.T) {
	r, err := cloneRepo("https://github.com/thundersparkf/adlnet-lrs-random-incorrect-url.git")
	assert.NotEmpty(t, err, "readGit_files.go :: cloneRepo :: err is nil")
	assert.Empty(t, r, "readGit_files.go :: cloneRepo :: repository is not nil")
	assert.Contains(t, err.Error(), errorHandlers.CloningErrorPrefix)

	empty, err := cloneRepo("")
	assert.Empty(t, empty, "readGit_files.go :: cloneRepo :: repository is not nil")
	assert.Contains(t, err.Error(), errorHandlers.CloningErrorPrefix)

}

func TestHappyGetTags(t *testing.T) {
	r, err := cloneRepo("https://github.com/thundersparkf/adlnet-lrs-py3.git")
	assert.NotEmpty(t, r, "readGit_files.go :: getTags :: r is nil")
	assert.Empty(t, err, "readGit_files.go :: getTags :: err is not nil")
	latestTags := getTags(r, "0.0.1")
	assert.NotEmpty(t, latestTags)
	assert.Contains(t, latestTags, "0.1.0")
	assert.Contains(t, latestTags, "0.2.1")
}

func TestUnhappyGetTags(t *testing.T) {
	r, err := cloneRepo("https://github.com/thundersparkf/adlnet-lrs-py3.git")
	assert.NotEmpty(t, r, "readGit_files.go :: getTags :: r is nil")
	assert.Empty(t, err, "readGit_files.go :: getTags :: err is not nil")
	moreLatestTags := getTags(r, "0.4.1")
	assert.Empty(t, moreLatestTags)

	empty, err := cloneRepo("")
	assert.NotEmpty(t, err, "readGit_files.go :: getTags :: err is nil")
	assert.Empty(t, empty, "readGit_files.go :: getTags :: r is not nil")
	latestTags := getTags(r, "0.4.1")
	assert.Empty(t, latestTags)
}

func TestGetSemverGreaterThanCurrent(t *testing.T) {
	assert.Equal(t, true, getSemverGreaterThanCurrent("1.0.0", "1.0.1"))
	assert.Equal(t, false, getSemverGreaterThanCurrent("1.0.0", "0.0.1"))
	assert.Equal(t, true, getSemverGreaterThanCurrent("1.0.0-alpha", "1.0.0"))
	assert.Equal(t, true, getSemverGreaterThanCurrent("1.0.0-alpha", "1.0.0-beta"))
	assert.Equal(t, false, getSemverGreaterThanCurrent("chaos", "1.0.0-beta"))
	assert.Equal(t, false, getSemverGreaterThanCurrent("1.0.0-beta", "chaos"))

}

func TestProcessGitRepo(t *testing.T) {
	_, updatedTags, err := processGitRepo("https://github.com/Darth-Tech/terraform-modules.git", "v1.0.2")
	assert.Empty(t, err, "readGitFiles :: processGitRepo :: error is not empty")
	assert.Equal(t, "v1.0.3-beta", updatedTags)

	_, multipleTags, err := processGitRepo("https://github.com/Darth-Tech/terraform-modules.git", "v1.0.1")
	assert.Empty(t, err, "readGitFiles :: processGitRepo :: error is not empty")
	assert.Contains(t, multipleTags, "v1.0.2")
	assert.Contains(t, multipleTags, "v1.0.3-beta")

	_, empty, err := processGitRepo("", "v1.0.1")
	assert.Empty(t, empty, "readGitFiles :: processGitRepo :: error is not empty")
	assert.NotEmpty(t, err)
}
