package cmd

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/thundersparkf/samwise/cmd/errorHandlers"
	"github.com/thundersparkf/samwise/cmd/outputs"
	"github.com/zclconf/go-cty/cty"
)

var FilesWritten []string

type reportJson struct {
	Report []jsonReport `json:"report"`
}
type jsonReport struct {
	RepoLink         string `json:"repo"`
	CurrentVersion   string `json:"current_version,omitempty"`
	UpdatesAvailable string `json:"updates_available,omitempty"`
	LatestVersion    string `json:"latest_update,omitempty"`
	FileName         string `json:"file_name"`
	Error            string `json:"error,omitempty"`
}

func Check(err error, message string, args ...any) {
	if err != nil {

		log.Error().Msgf(message, "errorArgs", args)
		panic(err)
	}
}

func CheckNonPanic(err error, message string, args ...any) bool {
	if err != nil {
		log.Error().Msgf(message, "errorArgs", args)
		log.Error().Msgf(err.Error())
		return true
	}
	return false
}

func cleanUpSourceString(source string) string {
	source = strings.ReplaceAll(source, "\"", "")
	source = strings.ReplaceAll(source, " ", "")
	return source
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
	log.Debug().Msgf("creating " + path + "/" + filename + ".csv file")
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
	headers := []string{"repo", "current_version", "file_name"}
	if LatestVersion {
		headers = append(headers, "latest_update")
	} else {
		headers = append(headers, "updates_available")
	}
	err = writer.Write(headers)
	Check(err, "unable to write headers to file", reportFilePath)
	for _, row := range data {
		if len(row["updates_available"]) > 0 || row["latest_update"] != "" {
			if LatestVersion {
				err = writer.Write([]string{row["repo"], row["current_version"], row["file_name"], row["latest_update"]})
			} else {
				err = writer.Write([]string{row["repo"], row["current_version"], row["file_name"], row["updates_available"]})
			}
			Check(err, "util :: CreateCSVReportFile :: unable to write record to file", row["repo"], row["current_version"], row["updates_available"], row["file_name"])
			writer.Flush()
		}
	}
	log.Debug().Msgf("created " + reportFilePath)

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
	log.Debug().Msgf("util :: createJSONReportFile :: reportString :: " + string(reportString))
	Check(err, "util :: createJSONReportFile :: unable unmarshal into output format")
	var nonEmptyRecords []jsonReport
	for _, value := range reportJsonObject {
		if value.CurrentVersion != "" {
			nonEmptyRecords = append(nonEmptyRecords, value)
		}
	}
	finalReportMap := map[string][]jsonReport{"report": nonEmptyRecords}

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
				log.Debug().Msgf("util :: readTfFiles :: sourceString :: %s", moduleSource)
				sources = append(sources, moduleSource)
			}
		}
	}
	return sources
}

func updateTfFiles(path string, fileName string) []string {
	log.Debug().Msgf("util :: updateTfFiles :: starting :: " + time.DateOnly)
	fullPath := path + "/" + fileName
	var sources = make([]string, 0)

	log.Debug().Msgf("util :: updateTfFiles :: reading file path :: %s", fullPath)
	content, _ := os.ReadFile(fullPath)
	file, _ := hclwrite.ParseConfig(content, fullPath, hcl.Pos{Line: 1, Column: 1})
	if file == nil {
		return []string{}
	}
	for _, block := range file.Body().Blocks() {
		labels := block.Labels()
		if block.Type() == "module" && len(labels) > 0 {
			log.Debug().Msgf("util :: updateTfFiles :: module detected")
			if block.Body().GetAttribute("source") != nil {
				log.Debug().Msgf("util :: updateTfFiles :: module source detected")
				sourceString := block.Body().GetAttribute("source").Expr().BuildTokens(nil).Bytes()
				moduleSource := string(sourceString)
				moduleSource = cleanUpSourceString(moduleSource)
				log.Debug().Msgf("util :: updateTfFiles :: sourceString :: %s", moduleSource)
				moduleSource = extractModuleSource(moduleSource)
				log.Debug().Msgf("util :: updateTfFiles :: moduleSource :: %s", moduleSource)
				if moduleSource != "" {
					sourceUrl, refTag, _ := extractRefAndPath(moduleSource)
					if refTag == "" {
						continue
					}
					log.Debug().Msgf("util :: updateTfFiles :: module data :: sourceUrl :: %s :: tag :: %s ", sourceUrl, refTag)
					_, tagsList, _ := processGitRepo(sourceUrl, refTag)
					largestTag := getGreatestSemverFromList(tagsList)
					if largestTag == "" {
						continue
					}
					sources = append(sources, fullPath)
					log.Debug().Msgf("util :: updateTfFiles :: file to be updated :: %s", fileName)
					log.Debug().Msgf("util :: updateTfFiles :: tag to updated :: currentSource:: %s :: tag :: %s :: tagList :: %s", moduleSource, largestTag, tagsList)
					currentSourceString := strings.Replace(moduleSource, refTag, largestTag, 1)
					log.Debug().Msgf("util :: updateTfFiles :: updatedModule :: %s", currentSourceString)
					writeHclBlockToFile(file, block, fullPath, "source", currentSourceString)
				}
			}
		}
	}
	log.Debug().Msgf("util :: updateTfFiles :: sources :: %s", sources)

	return sources
}

func writeHclBlockToFile(file *hclwrite.File, block *hclwrite.Block, path string, attr string, value any) {
	writeAttr := block.Body().SetAttributeValue(attr, cty.StringVal(value.(string)))
	log.Debug().Msgf("written to file :: writeAttr :: %s", string(writeAttr.Expr().BuildTokens(nil).Bytes()))
	writeFile, err := os.OpenFile(path, os.O_WRONLY, os.ModePerm)
	Check(err, "util :: updateTfFiles :: open file error")
	_, err = writeFile.Write(file.Bytes())
	Check(err, "util :: updateTfFiles :: write to file error")
	log.Debug().Msgf("util :: updateTfFiles :: path of output :: %s :: file_bytes :: %s", writeFile.Name(), string(file.Bytes()))
	err = writeFile.Close()
	CheckNonPanic(err, "util :: updateTfFiles :: unable to close file")
	log.Debug().Msgf("util :: updateTfFiles :: file closed")

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
	log.Debug().Msgf("util :: writeCommit :: opening repo")

	repo, err := git.PlainOpen(repoPath)
	CheckNonPanic(err, "util :: writeCommit :: failed to open repo")

	//repo, err := git.Open(memory., memfs.New())
	if CheckNonPanic(err, "util :: writeCommit :: unable to open repo") {
		return err
	}
	currentTime := time.Now()
	y, m, d := currentTime.Date()
	branch := "upgrade/tf-modules-" + fmt.Sprintf("%d-%d-%d", y, m, d)
	//err = repo.CreateBranch(&config.Branch{
	//	Name: "upgrade/tf-modules-" + time.DateOnly,
	//})

	if CheckNonPanic(err, "util :: writeCommit :: unable to create upgrade/tf-modules-"+time.DateOnly) {
		return err
	}

	w, err := repo.Worktree()
	Check(err, "util :: writeCommit :: worktree not fetched")
	log.Debug().Msgf("util :: writeCommit :: branch :: %s", branch)

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
		Check(err, "util :: writeCommit :: failed checkout of branch")

	}
	_, err = w.Add(".")
	Check(err, "util :: writeCommit :: unable to add files")

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

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func setupTerraform(workingDir string) *tfexec.Terraform {

	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.0.6")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Error().Msgf("error installing Terraform: ", "err", err)
		return nil
	}

	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Error().Msgf("error running NewTerraform: ", "err", err)
		return nil
	}
	return tf
}
