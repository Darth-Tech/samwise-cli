package cmd

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/thundersparkf/samwise/cmd/errorHandlers"
	"github.com/thundersparkf/samwise/cmd/outputs"
	"github.com/zclconf/go-cty/cty"
)

var FilesWritten []string

type reportJson struct {
	Report []jsonReport `json:"report"`
}
type jsonReport struct {
	RepoLink         string `json:"repo_Link"`
	CurrentVersion   string `json:"current_version"`
	UpdatesAvailable string `json:"updates_available"`
	Error            string `json:"error"`
}

func Check(err error, message string, args ...any) {
	if err != nil {

		slog.Error(message, slog.Any("errorArgs", args))
		panic(err)
	}
}

func CheckNonPanic(err error, message string, args ...any) bool {
	if err != nil {
		slog.Error(message, slog.Any("errorArgs", args))
		slog.Error(err.Error())
		return true
	}
	return false
}
func generateReport(data []map[string]string, outputFilename string, outputFormat string, path string) {
	if outputFormat == outputs.CSV {
		createCSVReportFile(data, path, outputFilename)
	} else if outputFormat == outputs.JSON {
		createJSONReportFile(data, path, outputFilename)
	} else {
		Check(errors.New("output format "+outputFormat+"not available"), "")
	}

}

func createCSVReportFile(data []map[string]string, path string, filename string) {
	slog.Debug("creating " + path + "/" + filename + ".csv file")
	reportFilePath := path + "/" + filename + ".csv"
	report, err := os.Create(reportFilePath)
	Check(err, "util :: createCSVReportFile :: unable to create file ", reportFilePath)
	defer func(report *os.File) {
		err := report.Close()
		if err != nil {
			Check(err, "util :: createCSVReportFile :: unable to close file")
		}
	}(report)

	writer := csv.NewWriter(report)
	defer writer.Flush()
	headers := []string{"repo_link", "current_version", "updates_available"}

	err = writer.Write(headers)
	Check(err, "unable to write headers to file", reportFilePath)
	for _, row := range data {
		if len(row["updates_available"]) > 0 {
			err := writer.Write([]string{row["repo"], row["current_version"], row["updates_available"]})
			Check(err, "util :: CreateCSVReportFile :: unable to write record to file", row["repo"], row["current_version"], row["updates_available"])
			writer.Flush()
		}
	}
	slog.Debug("created " + reportFilePath)

}

func checkOutputFormat(outputFormat string) (string, error) {
	outputFormat = strings.ToLower(outputFormat)
	outputsAvailable := []string{outputs.CSV, outputs.JSON}
	if !slices.Contains(outputsAvailable, outputFormat) {
		return "", errors.New(errorHandlers.CheckOutputFormatError)
	} else {
		return strings.ToLower(outputFormat), nil
	}
}

func checkOutputFilename(outputFilename string) string {
	extension := filepath.Ext(outputFilename)
	outputFilename = strings.ReplaceAll(outputFilename, extension, "")
	return outputFilename

}

func createJSONReportFile(data []map[string]string, path string, filename string) {
	reportFilePath := path + "/" + filename + ".json"
	report, err := os.Create(reportFilePath)
	Check(err, "unable to create file ", reportFilePath)
	defer func(report *os.File) {
		err := report.Close()
		if err != nil {
			Check(err, "util :: createJSONReportFile :: unable to close file")
		}
	}(report)
	reportString, err := json.Marshal(data)
	Check(err, "util :: createJSONReportFile :: unable to marshal modules data")
	var reportJsonObject []jsonReport
	err = json.Unmarshal(reportString, &reportJsonObject)
	slog.Debug("util :: createJSONReportFile :: reportString :: " + string(reportString))
	Check(err, "util :: createJSONReportFile :: unable unmarshal into output format")
	finalReportMap := map[string][]jsonReport{"report": reportJsonObject}
	reportOutputString, err := json.Marshal(finalReportMap)
	Check(err, "unable to marshal finalReportMap")
	_, err = report.Write(reportOutputString)
	Check(err, "util :: createJSONReportFile :: unable to write to file", reportFilePath)
}

func readTfFiles(path string) []string {
	var sources = make([]string, 0)
	content, _ := os.ReadFile(path)
	file, _ := hclwrite.ParseConfig(content, path, hcl.Pos{Line: 1, Column: 1})
	if file == nil {
		return []string{}
	}
	for _, block := range file.Body().Blocks() {
		labels := block.Labels()
		if block.Type() == "module" && len(labels) > 0 {
			if block.Body().GetAttribute("source") != nil {
				sourceString := block.Body().GetAttribute("source").Expr().BuildTokens(nil).Bytes()
				moduleSource := string(sourceString)
				moduleSource = strings.ReplaceAll(moduleSource, "\"", "")
				moduleSource = strings.ReplaceAll(moduleSource, " ", "")
				slog.Debug("util :: readTfFiles :: sourceString :: ", "sourceString", moduleSource)
				sources = append(sources, moduleSource)
			}
		}
	}
	return sources
}

func updateTfFiles(filepath string) []string {
	slog.Debug("util :: updateTfFiles :: starting :: " + time.DateOnly)
	var sources = make([]string, 0)

	slog.Debug("util :: updateTfFiles :: reading file", "filename", filepath)
	content, _ := os.ReadFile(filepath)
	file, _ := hclwrite.ParseConfig(content, filepath, hcl.Pos{Line: 1, Column: 1})
	if file == nil {
		return []string{}
	}
	for _, block := range file.Body().Blocks() {
		labels := block.Labels()
		if block.Type() == "module" && len(labels) > 0 {
			slog.Debug("util :: updateTfFiles :: module detected")
			if block.Body().GetAttribute("source") != nil {
				slog.Debug("util :: updateTfFiles :: module source detected")
				sourceString := block.Body().GetAttribute("source").Expr().BuildTokens(nil).Bytes()
				moduleSource := string(sourceString)
				moduleSource = strings.ReplaceAll(moduleSource, "\"", "")
				moduleSource = strings.ReplaceAll(moduleSource, " ", "")
				slog.Debug("util :: updateTfFiles :: sourceString :: ", "sourceString", moduleSource)
				moduleSource = extractModuleSource(moduleSource)
				slog.Debug("util :: updateTfFiles :: moduleSource :: ", "moduleSource", moduleSource)
				if moduleSource != "" {
					sourceUrl, refTag, _ := extractRefAndPath(moduleSource)
					if refTag == "" {
						continue
					}
					slog.Debug("util :: updateTfFiles :: module data", "sourceUrl", sourceUrl, "tag", refTag)
					_, tagsList, _ := processGitRepo(sourceUrl, refTag)
					largestTag := getGreatestSemverFromList(tagsList)
					if largestTag == "" {
						continue
					}
					slog.Debug("util :: updateTfFiles :: tag to updated :: ", "currentSource", string(moduleSource), "tag", largestTag, "tagList", tagsList)
					currentSourceString := strings.Replace(string(moduleSource), refTag, largestTag, 1)
					slog.Debug("util :: updateTfFiles :: tag to updated :: ", "currentSource", moduleSource, "tag", largestTag, "tagList", tagsList)

					writeAttr := block.Body().SetAttributeValue("source", cty.StringVal(currentSourceString))
					slog.Debug("written to file", "writeAttr", string(writeAttr.Expr().BuildTokens(nil).Bytes()))
					fmt.Printf("util :: updateTfFiles :: body :: %v", string(block.Body().BuildTokens(nil).Bytes()))
					writeFile, err := os.OpenFile(filepath, os.O_WRONLY, os.ModePerm)
					Check(err, "util :: updateTfFiles :: write to file error")
					_, err = writeFile.Write(file.Bytes())
					//fileInfo, err := writeFile.Stat()
					Check(err, "util :: updateTfFiles :: write to file error")
					slog.Debug("util :: updateTfFiles :: path of output", "output", writeFile.Name(), "file_bytes", string(file.Bytes()))
					slog.Debug("util :: updateTfFiles :: closing file", "filename", writeFile.Name())
					err = writeFile.Close()
					if !CheckNonPanic(err, "util :: updateTfFiles :: unable to close file") {
						FilesWritten = append(FilesWritten, filepath)
					}
					slog.Debug("util :: updateTfFiles :: writing to repo")
					writeCommit(Path)

				}

				sources = append(sources, moduleSource)
			}
		}
	}

	return sources
}

func getGreatestSemverFromList(tagsList string) string {
	if tagsList == "" {
		return ""
	}
	highestTag := "v0.0.0"
	tags := strings.Split(tagsList, "|")
	for _, tag := range tags {
		if getSemverGreaterThanCurrent(highestTag, tag) {
			highestTag = tag
		}
	}
	return highestTag
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

func writeCommit(repoPath string) error {
	slog.Debug("opening repo")
	repo, err := git.PlainOpen(repoPath)
	if CheckNonPanic(err, "util :: writeCommit :: unable to open repo") {
		return err
	}
	branch := "upgrade/tf-modules-" + time.DateOnly
	//err = repo.CreateBranch(&config.Branch{
	//	Name: "upgrade/tf-modules-" + time.DateOnly,
	//})

	if CheckNonPanic(err, "util :: writeCommit :: unable to create upgrade/tf-modules-"+time.DateOnly) {
		return err
	}

	w, err := repo.Worktree()

	slog.Debug("util :: writeCommit :: branch :: ", "branch", branch)

	branchRefName := plumbing.NewBranchReferenceName(branch)
	branchCoOpts := git.CheckoutOptions{
		Branch: plumbing.ReferenceName(branchRefName),
		Create: true,
	}

	if err := w.Checkout(&branchCoOpts); err != nil {
		//Check(err, fmt.Sprintf("local checkout of branch '%s' failed, will attempt to fetch remote branch of same name.", branch))

		/*mirrorRemoteBranchRefSpec := fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch)
		err = fetchOrigin(r, mirrorRemoteBranchRefSpec)
		CheckIfError(err)

		err = w.Checkout(&branchCoOpts)
		CheckIfError(err)
		*/
		branchCoOpts := git.CheckoutOptions{
			Branch: plumbing.ReferenceName(branchRefName),
			Create: false,
		}
		err = w.Checkout(&branchCoOpts)

	}
	_, err = w.Add(".")
	if CheckNonPanic(err, "util :: writeCommit :: unable to add") {
		return err
	}

	_, err = w.Status()
	if CheckNonPanic(err, "util :: writeCommit :: unable to fetch status") {
		return err
	}

	// Commits the current staging area to the repository, with the new file
	// just created. We should provide the object.Signature of Author of the
	// commit Since version 5.0.1, we can omit the Author signature, being read
	// from the git config files.
	commit, err := w.Commit("[updates] | updates the terraform modules to the latest version upstream", &git.CommitOptions{
		Author: &object.Signature{
			Name: "samwise",
			When: time.Now(),
		},
	})

	if CheckNonPanic(err, "util :: writeCommit :: unable to add") {
		return err
	}

	// Prints the current HEAD to verify that all worked well.
	_, err = repo.CommitObject(commit)
	if CheckNonPanic(err, "util :: writeCommit :: unable to add") {
		return err
	}
	//repo.Push(&git.PushOptions{})
	return nil
}
