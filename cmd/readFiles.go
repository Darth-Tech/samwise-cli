package cmd

import (
	"bufio"
	"golang.org/x/exp/maps"
	"log/slog"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var (
	// Regex to check if the line has "source=" in it
	sourceLineRegex      = regexp.MustCompile("source=\"(.+)\"")
	moduleSourceRegexMap = map[string]*regexp.Regexp{
		"generic_git": regexp.MustCompile("source=\"git::(.+)\""),
		"github":      regexp.MustCompile("source=\"(github.com.+)\""),
		"https":       regexp.MustCompile("source=\"(https://.+)\""),
		"bitbucket":   regexp.MustCompile("source=\".*(bitbucket.org.+)\""),
	}
	moduleRepoList []map[string]string
)

func fixTrailingSlashForPath(path string) string {
	if strings.HasSuffix(path, "/") {
		return strings.TrimRight(path, "/")
	}
	return path
}

func checkRegexMatchNotEmpty(match [][]string) string {
	if len(match) > 0 && len(match[0][1]) > 0 {
		return match[0][1]
	}
	return ""
}

func extractRefAndPath(sourceUrl string) (string, string) {
	urlParsed, err := url.Parse(sourceUrl)
	Check(err)
	params, err := url.ParseQuery(urlParsed.RawQuery)
	Check(err)
	var refTag string
	var rawUrl string
	if urlParsed.Scheme != "" {
		rawUrl = urlParsed.Scheme + "://"
	}
	rawUrl = rawUrl + urlParsed.Host + urlParsed.Path
	if params.Has("ref") {
		refTag = params.Get("ref")
		return rawUrl, refTag
	}
	return rawUrl, ""
}

func extractModuleSource(line string) string {
	keys := maps.Keys(moduleSourceRegexMap)
	var match [][]string
	var matchedString = ""
	for _, sourceRegex := range keys {
		match = moduleSourceRegexMap[sourceRegex].FindAllStringSubmatch(line, 1)
		matchedString = checkRegexMatchNotEmpty(match)
		if matchedString != "" {
			break
		}
	}
	return matchedString
}

func preProcessingSourceString(line string) (string, string) {
	// Will help avoid running moduleSourceRegexMap on every string
	line = strings.ReplaceAll(line, " ", "")
	sourceLineCheck := sourceLineRegex.FindAllStringSubmatch(line, -1)
	if checkRegexMatchNotEmpty(sourceLineCheck) == "" {
		return "", ""
	} else {
		repoLink := extractModuleSource(line)
		slog.Debug("Git repo link before: " + repoLink)
		var sourceUrl, refTag string
		if repoLink != "" {
			sourceUrl, refTag = extractRefAndPath(repoLink)
		}
		slog.Debug("Git repo link after: " + sourceUrl)
		return sourceUrl, refTag
	}
}

func processRepoLinksAndTags(path string) []map[string]string {
	files, err := os.ReadDir(fixTrailingSlashForPath(path))
	Check(err)
	for _, file := range files {
		fullPath := path + "/" + file.Name()
		f, err := os.Open(fullPath)
		Check(err)
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			repo, tag := preProcessingSourceString(line)
			slog.Debug("Repo: " + repo)
			if repo != "" {
				moduleRepoList = append(moduleRepoList, map[string]string{"repo": repo, "current_version": tag})
			}

		}
		err = f.Close()
		Check(err)
	}
	return moduleRepoList
}
