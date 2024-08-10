package cmd

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hashicorp/go-version"
	"github.com/spf13/viper"
	"github.com/thundersparkf/samwise/cmd/errorHandlers"
	"log/slog"
	"strings"
) // with go modules enabled (GO111MODULE=on or outside GOPATH)

func cloneRepo(url string) (*git.Repository, error) {
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: url,
		Auth: &http.BasicAuth{
			Username: viper.GetString("github_username"),
			Password: viper.GetString("github_key"),
		},
	})
	if err != nil {
		slog.Error("url: " + url)
		return nil, errors.New(errorHandlers.CloningErrorPrefix + err.Error())
	}
	return r, nil
}

func getTags(r *git.Repository, currentVersionTag string) string {
	tags, err := r.Tags()
	var tagsList []string
	Check(err)
	err = tags.ForEach(func(t *plumbing.Reference) error {
		versionToCheck := strings.ReplaceAll(t.Name().String(), "refs/tags/", "")
		if getSemverGreaterThanCurrent(currentVersionTag, versionToCheck) {

			tagsList = append(tagsList, versionToCheck)
		}
		return nil
	})
	if len(tagsList) > 0 {
		return strings.Join(tagsList, "|")
	}
	return ""
}

func getSemverGreaterThanCurrent(currentVersion string, versionToCheck string) bool {
	currentVersionTag, err := version.NewVersion(currentVersion)
	if err != nil {
		return false
	}
	versionToCheckTag, err := version.NewVersion(versionToCheck)
	if err != nil {
		return false
	}
	if versionToCheckTag.GreaterThan(currentVersionTag) {
		return true
	}
	return false

}
func processGitRepo(url string, currentVersionTag string) (string, error) {
	repo, err := cloneRepo(url)
	if repo != nil {
		tagsList := getTags(repo, currentVersionTag)
		return tagsList, nil
	}
	return "", err
}
