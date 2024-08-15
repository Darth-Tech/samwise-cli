package cmd

import (
	"bufio"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"golang.org/x/exp/maps"
)

var (
	// Regex to check if the line has "source=" in it
	sourceLineRegex      = regexp.MustCompile(`source="(.+\..+)"`)
	submoduleRegex       = regexp.MustCompile(`(?P<base_url>.*/.*)//(?P<submodule>.*)`)
	moduleSourceRegexMap = map[string]*regexp.Regexp{
		"generic_git": regexp.MustCompile(`git::(.+)`),
		"github":      regexp.MustCompile(`(.*github.com.+)`),
		"https":       regexp.MustCompile(`(https://.+)`),
		"bitbucket":   regexp.MustCompile(`.*(bitbucket.org.+)`),
	}
	removeUrlParams = regexp.MustCompile(`(\?.*)`)
	refRegex        = regexp.MustCompile(".*?ref=(.*)&.*|.*?ref=(.*)")
	moduleRepoList  []map[string]string
)

func fixTrailingSlashForPath(path string) string {
	if strings.HasSuffix(path, "/") {
		return strings.TrimRight(path, "/")
	}
	return path
}

func getNamedMatchesForRegex(reg *regexp.Regexp, sourceString string) map[string]string {
	match := reg.FindStringSubmatch(sourceString)
	result := make(map[string]string)

	if len(match) > 0 {
		for i, name := range reg.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}
	}

	return result
}

// Returns base url and cleaned paths
func extractSubmoduleFromSource(source string) (string, string) {
	subModuleMatch := getNamedMatchesForRegex(submoduleRegex, source)
	var baseUrl = ""
	var path = ""
	// If there is a match and one group is matched(1st), then it
	if len(subModuleMatch) > 0 && len(subModuleMatch["base_url"]) > 0 {
		baseUrl = subModuleMatch["base_url"]
	} else {
		baseUrl = source
	}
	if len(subModuleMatch) > 0 && len(subModuleMatch["submodule"]) > 0 {
		path = subModuleMatch["submodule"]
	}
	return baseUrl, path
}

func getTagFromUrl(source string) string {
	var refTag string
	refTagMatches := refRegex.FindStringSubmatch(source)
	if len(refTagMatches) > 0 {
		refTag = refTagMatches[2]
		return refTag
	}

	return refTag
}

// Returns url, tag submodules(if any) in that order
func extractRefAndPath(sourceUrl string) (string, string, string) {
	var refTag, submodulePaths string

	baseUrl, submodulePathsParams := extractSubmoduleFromSource(sourceUrl)
	submodulePaths = removeUrlParams.ReplaceAllString(submodulePathsParams, "")
	baseUrl = removeUrlParams.ReplaceAllString(baseUrl, "")
	refTag = getTagFromUrl(sourceUrl)

	return baseUrl, refTag, submodulePaths
}

func extractModuleSource(line string) string {
	var matchedString = ""
	match := sourceLineRegex.FindStringSubmatch(line)
	if len(match) > 0 {
		matchedString = match[1]
		if strings.Contains(matchedString, "@") {
			matchedString = strings.ReplaceAll(matchedString, "git::", "")
			return matchedString
		}
		for _, regex := range maps.Keys(moduleSourceRegexMap) {
			source := moduleSourceRegexMap[regex].FindStringSubmatch(matchedString)
			if len(source) > 0 {
				matchedString = strings.ReplaceAll(source[1], "git::", "")
				break
			}
		}

	}
	return matchedString
}

// Returns url, tag submodules(if any) in that order
func preProcessingSourceString(line string) (string, string, string) {
	// Will help avoid running moduleSourceRegexMap on every string
	line = strings.ReplaceAll(line, " ", "")
	sourceLineCheck := sourceLineRegex.FindStringSubmatch(line)
	if len(sourceLineCheck) == 0 {
		return "", "", ""
	} else {
		repoLink := extractModuleSource(line)
		//repoLink := sourceLineCheck[1]
		slog.Debug("Git repo link before: " + repoLink)
		var sourceUrl, refTag, submodule string
		if repoLink != "" {

			sourceUrl, refTag, submodule = extractRefAndPath(repoLink)
		}
		slog.Debug("Git repo link after: " + sourceUrl)
		return sourceUrl, refTag, submodule
	}
}

func processRepoLinksAndTags(path string) []map[string]string {
	files, err := os.ReadDir(fixTrailingSlashForPath(path))
	if CheckNonPanic(err, "readFiles :: processRepoLinksAndTags :: unable to read directory", path) {
		return nil
	}
	for _, file := range files {
		fullPath := path + "/" + file.Name()
		f, err := os.Open(fullPath)
		if CheckNonPanic(err, "readFiles :: processRepoLinksAndTags :: unable to read file", path, fullPath) {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			repo, tag, submodule := preProcessingSourceString(line)
			slog.Debug("readFiles :: processRepoLinksAndTags :: repo url :: " + repo)
			if repo != "" {
				moduleRepoList = append(moduleRepoList, map[string]string{"repo": repo, "current_version": tag, "submodule": submodule})
			}

		}
		err = f.Close()
		if CheckNonPanic(err, "readFiles :: processRepoLinksAndTags :: unable to close file", path, fullPath) {
			continue
		}
	}
	return moduleRepoList
}
