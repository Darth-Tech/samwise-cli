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
	submoduleRegex  = regexp.MustCompile("(.*/.*)//(.*)")
	removeUrlParams = regexp.MustCompile("(\\?.*)")
	moduleRepoList  []map[string]string
)

func fixTrailingSlashForPath(path string) string {
	if strings.HasSuffix(path, "/") {
		return strings.TrimRight(path, "/")
	}
	return path
}

// Returns base url and cleaned paths
func extractSubmoduleFromSource(source string) (string, string) {
	subModuleMatch := submoduleRegex.FindAllStringSubmatch(source, -1)
	var baseUrl = ""
	var path = ""
	if len(subModuleMatch) > 0 && len(subModuleMatch[0][1]) > 0 {
		baseUrl = subModuleMatch[0][1]
	} else {
		baseUrl = source
	}
	if len(subModuleMatch) > 0 && len(subModuleMatch[0][2]) > 0 {
		path = subModuleMatch[0][2]
	}
	return baseUrl, path
}
func checkRegexMatchNotEmpty(match [][]string) string {
	if len(match) > 0 && len(match[0][1]) > 0 {
		return match[0][1]
	}
	return ""
}

// TODO Add tests
func getTagFromUrl(source string) string {
	var refTag string
	refParams, err := url.Parse(source)
	// TODO: Refactor
	if CheckNonPanic(err, "readFiles :: extractRefAndPath :: unable to parse url for params") {
		refTag = ""
	} else {
		params, err := url.ParseQuery(refParams.RawQuery)
		if CheckNonPanic(err, "readFiles :: extractRefAndPath :: unable to parse url for params") {
			refTag = ""
		}
		refTag = params.Get("ref")
	}
	return refTag
}

// Returns url, tag submodules(if any) in that order
func extractRefAndPath(sourceUrl string) (string, string, string) {
	var refTag, finalUrl, tempUrl, submodulePaths, submodulePathsParams string
	if strings.Count(sourceUrl, "//") == 2 {
		sourceUrl, submodulePathsParams = extractSubmoduleFromSource(sourceUrl)
		refTag = getTagFromUrl(sourceUrl + "/" + submodulePathsParams)
		submodulePaths = removeUrlParams.ReplaceAllString(submodulePathsParams, "")
	} else {
		refTag = getTagFromUrl(sourceUrl)
	}
	tempUrl = sourceUrl
	urlCleaner, _ := url.Parse(tempUrl)

	if urlCleaner.Scheme != "" {
		finalUrl = urlCleaner.Scheme + "://"
	}
	finalUrl = finalUrl + urlCleaner.Host + urlCleaner.Path

	return finalUrl, refTag, submodulePaths
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

// Returns url, tag submodules(if any) in that order
func preProcessingSourceString(line string) (string, string, string) {
	// Will help avoid running moduleSourceRegexMap on every string
	line = strings.ReplaceAll(line, " ", "")
	sourceLineCheck := sourceLineRegex.FindAllStringSubmatch(line, -1)
	if checkRegexMatchNotEmpty(sourceLineCheck) == "" {
		return "", "", ""
	} else {
		repoLink := extractModuleSource(line)
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
