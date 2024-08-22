//coverage:ignore
/*
Copyright © 2024 Agastya Dev Addepally (devagastya0@gmail.com)
*/
package cmd

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var modulesListTotal []map[string]string
var failureListTotal []map[string]string
var Path string
var Verbose bool
var OutputFormat string
var OutputFilename string
var Depth int

// checkForUpdatesCmd represents the checkForUpdates command
var checkForUpdatesCmd = &cobra.Command{
	Use:   "checkForUpdates --path=[Directory with module usage]",
	Short: "search for updates for terraform modules using in your code and generate a report",
	Long: `
	
	Searches (sub)directories for module sources and versions to create a report listing versions available for updates.

CSV format : repo_link | current_version | updates_available

JSON format: [{
                "repo_link": <repo_link>,
                "current_version": <current version used in the code>,
                "updates_available"
             }]

An update is never late, nor is it early, it arrives precisely when it means to.
	`,

	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("creating a report..."+Path, "verbose", Verbose)
		_, _, directoriesToIgnore, _, _ := getParamsForCheckForUpdatesCMD(cmd.Flags())
		slog.Debug("output format: " + OutputFormat)
		slog.Debug("Params: ", slog.String("depth", strconv.Itoa(Depth)), slog.String("rootDir", Path), slog.String("directoriesToIgnore", strings.Join(directoriesToIgnore, " ")))
		rootDir := fixTrailingSlashForPath(Path)
		var modules []map[string]string
		var failureList []map[string]string
		err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
			Check(err, "checkForUpdates :: command :: ", path)
			depthCountInCurrentPath := strings.Count(rootDir, string(os.PathSeparator))
			if d.IsDir() && !slices.Contains(directoriesToIgnore, d.Name()) {
				slog.Debug("checkForUpdates :: command :: in directory " + path)
				if strings.Count(path, string(os.PathSeparator)) > depthCountInCurrentPath+Depth {
					slog.Debug("...which is skipped")
					return fs.SkipDir
				}
				modules, failureList = checkForModuleSourceUpdates(path)
				slog.Debug("checkForUpdates :: command :: ", "modules", modules)
				modulesListTotal = append(modulesListTotal, modules...)
				failureListTotal = append(failureListTotal, failureList...)
			}
			return nil
		})
		Check(err, "checkForUpdates :: command :: unable to walk the directories")
		OutputFormat, err = checkOutputFormat(OutputFormat)
		Check(err, "checkForUpdates :: command :: output format error", OutputFormat)
		OutputFilename = checkOutputFilename(OutputFilename)
		generateReport(modulesListTotal, OutputFilename, OutputFormat, rootDir)
		createJSONReportFile(failureList, rootDir, "failure_report")

	},
}

func checkForModuleSourceUpdates(path string) ([]map[string]string, []map[string]string) {
	var modules []map[string]string
	var failureList []map[string]string
	var listWritten []string
	var bar *progressbar.ProgressBar
	path = fixTrailingSlashForPath(path)
	modules = processRepoLinksAndTags(path)
	slog.Debug("checkForUpdates :: checkForModuleSourceUpdates :: path: " + path)


	slog.Info("Scanning directory " + path + " ...")
	if len(modules) > 0 {
		bar = progressbar.Default(int64(len(modules)))
	}
	for _, module := range modules {
		err := bar.Add(1)
		Check(err, "progressbar error")
		if !slices.Contains(listWritten, module["repo"]) {
			_, tagsList, err := processGitRepo(module["repo"], module["current_version"])
			if err != nil {
				failureList = append(failureList, map[string]string{
					"repo":              module["repo"],
					"current_version":   module["current_version"],
					"updates_available": tagsList,
					"error":             err.Error(),
				})
			}
			if len(tagsList) > 0 {
				module["updates_available"] = tagsList
				listWritten = append(listWritten, module["repo"])
			}
			slog.Debug("checkForUpdates :: checkForModuleSourceUpdates :: path :: ", "repo", module["repo"], "current", module["current_version"], "updates_available", module["updates_available"])

		}
	}

	return modules, failureList
}

// Fixed return of params depth, rootDir, directoriesToIgnore, output, outputFilename
func getParamsForCheckForUpdatesCMD(flags *pflag.FlagSet) (int, string, []string, string, string) {
	depth, err := flags.GetInt("depth")
	Check(err, "checkForUpdates :: command :: depth argument error")
	rootDir, err := flags.GetString("path")
	Check(err, "checkForUpdates :: command :: path argument error")
	directoriesToIgnore, err := flags.GetStringArray("ignore")
	Check(err, "checkForUpdates :: command :: ignore argument error")
	output, err := flags.GetString("output")
	Check(err, "checkForUpdates :: command :: output argument error")
	outputFilename, err := flags.GetString("output-filename")
	Check(err, "checkForUpdates :: command :: output-filename argument error")
	return depth, rootDir, directoriesToIgnore, output, outputFilename
}

func init() {
	cobra.OnInitialize(initConfig)
	//checkForUpdatesCmd.Flags().Bool("ci", false, "Set this flag for usage in CI systems. Does not generate a report. Prints JSON to Stdout and returns exit code 1 if modules are outdated.")
	//checkForUpdatesCmd.Flags().Bool("allow-failure", true, "Set this flag for usage in CI systems. If true, does NOT exit code 1 when modules are outdated.")
	rootCmd.AddCommand(checkForUpdatesCmd)

	checkForUpdatesCmd.PersistentFlags().IntVarP(&Depth, "depth", "d", 0, "Folder depth to search for modules in. Give -1 for a full directory extraction.")
	checkForUpdatesCmd.PersistentFlags().StringVar(&Path, "path", "p", "The path for directory containing terraform code to extract modules from.")
	checkForUpdatesCmd.PersistentFlags().String("git-repo", "g", "Git Repository to check module dependencies on.")
	checkForUpdatesCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "The path for directory containing terraform code to extract modules from.")
	checkForUpdatesCmd.PersistentFlags().StringArrayP("ignore", "i", []string{".git", ".idea"}, "Directories to ignore when searching for the One Ring(modules and their sources.")
	checkForUpdatesCmd.PersistentFlags().StringVarP(&OutputFormat, "output", "o", "csv", "Output format. Supports \"csv\" and \"json\". Default value is csv.")
	checkForUpdatesCmd.PersistentFlags().StringVarP(&OutputFilename, "output-filename", "f", "module_report", "Output file name.")

	err := checkForUpdatesCmd.MarkPersistentFlagRequired("path")
	if err != nil {
		return
	}

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkForUpdatesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

}
