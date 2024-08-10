/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
		slog.Debug("creating a CSV report...")
		depth, rootDir, directoriesToIgnore := getParamsForCheckForUpdatesCMD(cmd.Flags())
		slog.Debug("Params: ", strconv.Itoa(depth), rootDir, directoriesToIgnore)
		rootDir = fixTrailingSlashForPath(rootDir)
		var modules []map[string]string
		var listWritten []string
		slog.Debug("creating " + rootDir + "/module_dependency_report.csv file")
		f, writer := CreateCSVReportFile(rootDir + "/module_dependency_report.csv")
		slog.Debug("created " + rootDir + "/module_dependency_report.csv file")
		defer f.Close()
		_, err := writer.WriteString("repo_link,current_version,tag_list\n")
		Check(err)
		err = writer.Flush()
		Check(err)
		err = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
			Check(err)
			depthCountInCurrentPath := strings.Count(rootDir, string(os.PathSeparator))
			if d.IsDir() && !slices.Contains(directoriesToIgnore, d.Name()) {
				slog.Debug("In directory " + path + "...")
				if strings.Count(path, string(os.PathSeparator)) > depthCountInCurrentPath+depth {
					slog.Debug("...which is skipped")
					return fs.SkipDir
				}
				path = fixTrailingSlashForPath(path)
				modules = processRepoLinksAndTags(path)
				bar := progressbar.Default(int64(len(modules)))
				for _, module := range modules {
					bar.Add(1)
					slog.Debug(module["repo"])
					if !slices.Contains(listWritten, module["repo"]) {
						tagsList := processGitRepo(module["repo"], module["current_version"])
						if len(tagsList) > 0 {
							writer.WriteString(module["repo"] + "," + module["current_version"] + "," + tagsList + "\n")
							writer.Flush()
							listWritten = append(listWritten, module["repo"])
						}
					}
				}
			}
			return nil
		})
		Check(err)
	},
}

func getParamsForCheckForUpdatesCMD(flags *pflag.FlagSet) (int, string, []string) {
	depth, err := flags.GetInt("depth")
	Check(err)
	rootDir, err := flags.GetString("path")
	Check(err)
	directoriesToIgnore, err := flags.GetStringArray("ignore")
	Check(err)
	return depth, rootDir, directoriesToIgnore
}

func init() {
	cobra.OnInitialize(initConfig)
	checkForUpdatesCmd.Flags().IntP("depth", "d", 0, "Folder depth to search for modules in. Give -1 for a full directory extraction.")
	checkForUpdatesCmd.Flags().StringP("path", "p", ".", "The path for directory containing terraform code to extract modules from.")
	checkForUpdatesCmd.Flags().String("git-repo", "g", "Git Repository to check module dependencies on.")
	checkForUpdatesCmd.Flags().StringArrayP("ignore", "i", []string{}, "Directories to ignore when searching for the One Ring(modules and their sources.")

	rootCmd.AddCommand(checkForUpdatesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkForUpdatesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

}
