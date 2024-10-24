package cmd

import (
	"errors"
	"github.com/rs/zerolog/log"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	sshgit "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/spf13/viper"
	"github.com/thundersparkf/samwise/cmd/errorHandlers"
	"golang.org/x/crypto/ssh"
) // with go modules enabled (GO111MODULE=on or outside GOPATH)

func gitAuthGenerator(url string) transport.AuthMethod {
	if !strings.Contains(url, "@") {
		log.Debug().Msg("using basic https auth")
		return &http.BasicAuth{
			Username: viper.GetString("git_user"),
			Password: viper.GetString("git_key"),
		}
	}
	log.Debug().Msg("using ssh auth")
	username := strings.Split(url, "@")[0]
	sshPath := viper.GetString("git_ssh_key_path")
	log.Debug().Msg("readGitFiles :: gitAuthGenerator :: " + sshPath)
	sshKey, err := os.ReadFile(sshPath)
	Check(err, "filename "+sshPath)
	signer, err := ssh.ParsePrivateKey(sshKey)
	Check(err, "signer key died")
	publicKey := &sshgit.PublicKeys{User: username, Signer: signer, HostKeyCallbackHelper: sshgit.HostKeyCallbackHelper{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}}
	return publicKey
}

func parseGitUrl(source string) string {
	log.Debug().Msg("readGitFiles :: parseGitUrl :: source " + source)
	if strings.Contains(source, "@") || strings.Contains(source, "ssh://") {
		source = strings.Replace(source, "ssh://", "", 1)
		return source
	}
	source = strings.Replace(source, "git::", "", 1)
	endpointUrl, err := transport.NewEndpoint(source)
	log.Debug().Msgf("readGitFiles :: parseGitUrl :: endpoint result :: host :: %s :: path :: %s :: protocol :: %s :: ", endpointUrl.Host, endpointUrl.Path, endpointUrl.Protocol)
	if CheckNonPanic(err, "unable to parse git url") {
		return ""
	}

	if endpointUrl.Protocol == "" || endpointUrl.Protocol == "file" {
		return "https://" + strings.Replace(endpointUrl.String(), "file://", "", 1)
	}
	return endpointUrl.String()
}
func cloneRepo(url string) (*git.Repository, error) {
	url = parseGitUrl(url)
	log.Debug().Msg("readGitFiles :: cloneRepo :: url :: " + url)
	authMethod := gitAuthGenerator(url)
	//authMethod, err := ssh.DefaultAuthBuilder("git")
	if url == "" {
		log.Debug().Msg("readGitFiles :: cloneRepo :: url is empty from parseGitUrl")
		return nil, errors.New(errorHandlers.CloningErrorPrefix + " unable to clone " + url)
	}
	log.Debug().Msgf("readGitFiles :: cloneRepo :: auth method :: %s", authMethod.String())
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:  url,
		Auth: authMethod,
	})
	if err != nil {
		log.Debug().Msg("readGitFiles :: cloneRepo :: url :: " + url)
		return nil, errors.New(errorHandlers.CloningErrorPrefix + err.Error())
	}
	return r, nil
}

func getTags(r *git.Repository, currentVersionTag string) string {
	tags, err := r.Tags()
	var tagsList []string
	if err != nil {
		log.Error().Msg("readGitFiles :: getTags :: unable to get tags :: " + err.Error())
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
	if CheckNonPanic(err, "unable to retrieve tags") {
		return ""
	}
	if len(tagsList) > 0 {
		return strings.Join(tagsList, "|")
	}
	return ""
}

func processGitRepo(url string, currentVersionTag string) (*git.Repository, string, error) {
	repo, err := cloneRepo(url)
	if repo != nil {
		tagsList := getTags(repo, currentVersionTag)
		return repo, tagsList, nil
	}
	return nil, "", err
}
