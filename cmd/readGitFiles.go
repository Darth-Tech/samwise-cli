package cmd

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hashicorp/go-version"
	"github.com/spf13/viper"
	"log/slog"
	"strings"
) // with go modules enabled (GO111MODULE=on or outside GOPATH)

func cloneRepo(url string) *git.Repository {
	slog.Debug("username: " + viper.GetString("github_username"))
	slog.Debug("password: " + viper.GetString("github_key"))

	/*var progressBarSetting *os.File = nil
	if slog.LevelKey == slog.LevelDebug.String() {
		progressBarSetting = os.Stdout
	}*/
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: url,
		Auth: &http.BasicAuth{
			Username: viper.GetString("github_username"),
			Password: viper.GetString("github_key"),
		},
	})
	if err != nil {
		slog.Error("URL: " + url)
		//Check(err)
		return nil
	}
	return r
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
func processGitRepo(url string, currentVersionTag string) string {
	repo := cloneRepo(url)
	if repo != nil {
		tagsList := getTags(repo, currentVersionTag)
		//fmt.Println(tagsList)
		return tagsList
	}
	return ""
}
