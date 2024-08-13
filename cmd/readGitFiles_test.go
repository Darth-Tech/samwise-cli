package cmd

import (
	"github.com/stretchr/testify/assert"
	"github.com/thundersparkf/samwise/cmd/errorHandlers"
	"testing"
)

func TestGitAuthenticationGenerator(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.PanicsWithValue(t, "git_user value not set", func() {
				gitAuthGenerator()
			})
		}
	}()
}
func TestHappyClonePublicRepo(t *testing.T) {
	r, err := cloneRepo("https://github.com/thundersparkf/adlnet-lrs-py3.git")
	assert.NotEmpty(t, r, "readGit_files.go :: cloneRepo :: repository is nil")
	assert.Empty(t, err, "readGit_files.go :: cloneRepo :: err is not nil :: ")
	head, err := r.Head()
	assert.Equal(t, "refs/heads/master", head.Name().String(), "readGit_files.go :: cloneRepo :: repository's head is ref/heads/master")

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
