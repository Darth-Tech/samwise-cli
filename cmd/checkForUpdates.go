/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/thundersparkf/samwise/cmd/outputs"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// checkForUpdatesCmd represents the checkForUpdates command
var checkForUpdatesCmd = &cobra.Command{
	Use:   "checkForUpdates",
	Short: "search for updates for terraform modules using in your code and generate a report",
	Long: `Searches (sub)directories for module sources and versions to create a report listing versions available for updates.
CSV format : repo_link | current_version | versions_available
JSON format: [{
                "repo_link": <repo_link>,
                "current_version": <current version used in the code>,
                "versions_available"
             }]
	`,
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("creating a report...")
		depth, rootDir, directoriesToIgnore, outputFormat := getParamsForCheckForUpdatesCMD(cmd.Flags())
		slog.Debug("output format: " + outputFormat)
		slog.Debug("Params: ", slog.String("depth", strconv.Itoa(depth)), slog.String("rootDir", rootDir), slog.String("directoriesToIgnore", strings.Join(directoriesToIgnore, " ")))
		rootDir = fixTrailingSlashForPath(rootDir)
		var modules []map[string]string
		var listWritten []string
		err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
			Check(err)
			depthCountInCurrentPath := strings.Count(rootDir, string(os.PathSeparator))
			if d.IsDir() && !slices.Contains(directoriesToIgnore, d.Name()) {
				slog.Debug("In directory " + path)
				if strings.Count(path, string(os.PathSeparator)) > depthCountInCurrentPath+depth {
					slog.Debug("...which is skipped")
					return fs.SkipDir
				}
				path = fixTrailingSlashForPath(path)
				modules = processRepoLinksAndTags(path)
				bar := progressbar.Default(int64(len(modules)))
				slog.Debug("Path: " + path)
				for _, module := range modules {
					bar.Add(1)
					slog.Debug(module["repo"])
					if !slices.Contains(listWritten, module["repo"]) {
						tagsList, _ := processGitRepo(module["repo"], module["current_version"])
						if len(tagsList) > 0 {
							module["tags_available_for_update"] = tagsList
							listWritten = append(listWritten, module["repo"])
						}
					}
				}
			}
			return nil
		})
		Check(err)
		outputFormat, err = checkOutputFormat(outputFormat)
		Check(err)
		generateReport(modules, outputFormat, rootDir)
	},
}

func getParamsForCheckForUpdatesCMD(flags *pflag.FlagSet) (int, string, []string, string) {
	depth, err := flags.GetInt("depth")
	Check(err)
	rootDir, err := flags.GetString("path")
	Check(err)
	directoriesToIgnore, err := flags.GetStringArray("ignore")
	Check(err)
	output, err := flags.GetString("output")
	Check(err)
	return depth, rootDir, directoriesToIgnore, output
}

func generateReport(data []map[string]string, outputFormat string, path string) {
	if outputFormat == outputs.CSV {
		CreateCSVReportFile(data, path)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	checkForUpdatesCmd.Flags().IntP("depth", "d", 0, "Folder depth to search for modules in. Give -1 for a full directory extraction.")
	checkForUpdatesCmd.Flags().StringP("path", "p", ".", "The path for directory containing terraform code to extract modules from.")
	checkForUpdatesCmd.Flags().String("git-repo", "g", "Git Repository to check module dependencies on.")
	checkForUpdatesCmd.Flags().StringArrayP("ignore", "i", []string{".git", ".idea"}, "Directories to ignore when searching for the One Ring(modules and their sources.")
	checkForUpdatesCmd.Flags().StringP("output", "o", "csv", "Output format. Supports \"csv\" and \"json\". Default value is csv.")

	rootCmd.AddCommand(checkForUpdatesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkForUpdatesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

}
