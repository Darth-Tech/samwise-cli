package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thundersparkf/samwise/cmd/errorHandlers"
)

func TestGitAuthenticationGenerator(t *testing.T) {
	// TODO: Add tests to check key present and getting ssh type for urls with @
	// TODO: Add tests for key not present panic
	// TODO: Add tests for @ not present in url and username and password retrieved
	// TODO: Add tests for user and password not passed so they return empty strings
	// TODO: Incorrect format private key parsing panic
	defer func() {
		if r := recover(); r != nil {
			assert.PanicsWithValue(t, "git_user value not set", func() {
				gitAuthGenerator()
			})
		}
	}()
}

// TODO: Add tests for ssh cloning
func TestHappyClonePublicRepo(t *testing.T) {
	r, err := cloneRepo("https://github.com/thundersparkf/adlnet-lrs-py3.git")
	assert.NotEmpty(t, r, "readGit_files.go :: cloneRepo :: repository is nil")
	assert.Empty(t, err, "readGit_files.go :: cloneRepo :: err is not nil :: ")
	head, _ := r.Head()
	assert.Equal(t, "refs/heads/master", head.Name().String(), "readGit_files.go :: cloneRepo :: repository's head is ref/heads/master")

}

func TestUnhappyNoProtocolClonePublicRepo(t *testing.T) {
	//r, err := cloneRepo("github.com/thundersparkf/adlnet-lrs-py3.git")
	//assert.Empty(t, r, "readGit_files.go :: cloneRepo :: repository is nil")
	//assert.NotEmpty(t, err, "readGit_files.go :: cloneRepo :: err is not nil :: ")
	//head, err := r.Head()
	defer func() {
		if r := recover(); r != nil {
			//r, err :=
			assert.Panics(t, func() { cloneRepo("github.com/thundersparkf/adlnet-lrs-py3.git") }, "readGit_files.go :: cloneRepo :: repository's head is ref/heads/master")
		}
	}()

}

func TestUnhappyClonePublicRepo(t *testing.T) {
	r, err := cloneRepo("https://github.com/thundersparkf/adlnet-lrs-random-incorrect-url.git")
	assert.NotEmpty(t, err, "readGit_files.go :: cloneRepo :: err is nil")
	assert.Empty(t, r, "readGit_files.go :: cloneRepo :: repository is not nil")
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
}

func TestGetSemverGreaterThanCurrent(t *testing.T) {
	assert.Equal(t, true, getSemverGreaterThanCurrent("1.0.0", "1.0.1"))
	assert.Equal(t, false, getSemverGreaterThanCurrent("1.0.0", "0.0.1"))
	assert.Equal(t, true, getSemverGreaterThanCurrent("1.0.0-alpha", "1.0.0"))
	assert.Equal(t, true, getSemverGreaterThanCurrent("1.0.0-alpha", "1.0.0-beta"))
	assert.Equal(t, false, getSemverGreaterThanCurrent("chaos", "1.0.0-beta"))
}
