package cmd

import (
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	sshgit "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hashicorp/go-version"
	"github.com/spf13/viper"
	"github.com/thundersparkf/samwise/cmd/errorHandlers"
	"golang.org/x/crypto/ssh"
) // with go modules enabled (GO111MODULE=on or outside GOPATH)

func gitAuthGenerator(url string) transport.AuthMethod {
	if !strings.Contains(url, "@") {
		return &http.BasicAuth{
			Username: viper.GetString("git_user"),
			Password: viper.GetString("git_password"),
		}
	}
	username := strings.Split(url, "@")[0]
	sshPath := viper.Get("git_ssh_key_path")
	sshKey, err := os.ReadFile(sshPath.(string))
	Check(err, "key error")
	signer, err := ssh.ParsePrivateKey(sshKey)
	publicKey := &sshgit.PublicKeys{User: username, Signer: signer, HostKeyCallbackHelper: sshgit.HostKeyCallbackHelper{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}}

	if CheckNonPanic(err, "readGitFiles :: gitAuthGenerator :: unable to clone repo "+url) {
		return nil
	}
	return publicKey
}

func cloneRepo(url string) (*git.Repository, error) {
	authMethod := gitAuthGenerator(url)
	//authMethod, err := ssh.DefaultAuthBuilder("git")
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:  url,
		Auth: authMethod,
	})
	if err != nil {
		slog.Error("readGitFiles :: cloneRepo :: url: " + url)
		return nil, errors.New(errorHandlers.CloningErrorPrefix + err.Error())
	}
	return r, nil
}

func getTags(r *git.Repository, currentVersionTag string) string {
	tags, err := r.Tags()
	var tagsList []string
	if err != nil {
		slog.Error("readGitFiles :: getTags :: unable to get tags :: " + err.Error())
		return ""
	}
	//TODO: CheckNonPanic here
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
