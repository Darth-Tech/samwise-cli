package cmd

import (
	"bufio"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

var (
	moduleRepoList        []map[string]string
	findModuleSourceRegex = regexp.MustCompile("source=\"git::(.+)\"")
	refTagRegex           = regexp.MustCompile("(\\?ref=.+)")
)

func fixTrailingSlashForPath(path string) string {
	if strings.HasSuffix(path, "/") {
		return strings.TrimRight(path, "/")
	}
	return path
}

func checkRegexMatchNotempty(match [][]string) string {
	if len(match) > 0 && len(match[0][1]) > 0 {
		return match[0][1]
	}
	return ""

}

func preProcessingSourceString(line string) (string, string) {
	line = strings.ReplaceAll(line, " ", "")
	processedString := findModuleSourceRegex.FindAllStringSubmatch(line, -1)
	gitRepoLink := checkRegexMatchNotempty(processedString)
	refTag := refTagRegex.FindAllStringSubmatch(gitRepoLink, -1)
	tag := checkRegexMatchNotempty(refTag)
	if tag != "" {
		tag = strings.ReplaceAll(tag, "?ref=", "")
	}
	slog.Debug("Git repo link before: " + gitRepoLink)
	gitRepoLink = refTagRegex.ReplaceAllString(gitRepoLink, "")

	slog.Debug("Git repo link after: " + gitRepoLink)
	return gitRepoLink, tag
}

func processRepoLinksAndTags(path string) []map[string]string {
	files, err := os.ReadDir(fixTrailingSlashForPath(path))
	Check(err)
	for _, file := range files {
		fullPath := path + "/" + file.Name()
		file, err := os.Open(fullPath)
		Check(err)
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			repo, tag := preProcessingSourceString(line)
			// TODO: Shift left and avoid cloning duplicate
			if repo != "" {
				moduleRepoList = append(moduleRepoList, map[string]string{"repo": repo, "current_version": tag})
			}

		}
	}
	return moduleRepoList
}
