//coverage:ignore
/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var filesUpdatedTotal []string

// ciCmd represents the ci command
var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "For CI integrations[experimental]",
	Long: `
	
	Includes features for better CI integrations such as failure when updates available 
	for pipelines, allowing users to automatically create PRs when updates are present(custom thresholds) and so on.

Not all those who don't update dependencies are lost.`,
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("Ci stuff... in "+Path, "args", len(args))
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
		_, _, directoriesToIgnore, _, _ := getParamsForCheckForUpdatesCMD(cmd.Flags())
		slog.Debug("output format: " + OutputFormat)
		slog.Debug("Params: ", slog.String("depth", strconv.Itoa(Depth)), slog.String("rootDir", Path), slog.String("directoriesToIgnore", strings.Join(directoriesToIgnore, " ")))
		rootDir := fixTrailingSlashForPath(Path)
		tf := setupTerraform(rootDir)
		if tf == nil {
			return
		}
		var modules []map[string]string
		var failureList []map[string]string
		err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
			Check(err, "ci :: command :: ", path)
			depthCountInCurrentPath := strings.Count(rootDir, string(os.PathSeparator))
			if d.IsDir() && !slices.Contains(directoriesToIgnore, d.Name()) {
				slog.Debug("ci :: command :: in directory " + path)
				if strings.Count(path, string(os.PathSeparator)) > depthCountInCurrentPath+Depth {
					slog.Debug("...which is skipped")
					return fs.SkipDir
				}
				err := tf.FormatWrite(context.TODO())
				Check(err, "tf fmt failed")
				filesUpdated := createModuleVersionUpdates(path)
				filesUpdatedTotal = append(filesUpdatedTotal, filesUpdated...)
				filesUpdatedTotal = removeDuplicateStr(filesUpdatedTotal)
				modulesListTotal = append(modulesListTotal, modules...)
				failureListTotal = append(failureListTotal, failureList...)
			}
			return nil
		})
		Check(err, "ci :: command :: unable to walk the directories")
		slog.Debug("ci :: command :: filesUpdatedTotal", "filesUpdatedTotal", filesUpdatedTotal)
		writeCommit(rootDir)

	},
}

func createModuleVersionUpdates(path string) []string {
	files, err := os.ReadDir(fixTrailingSlashForPath(path))
	var filesUpdated []string
	Check(err, "util :: updateTfFiles :: unable to read dir")
	for _, file := range files {
		filesEdited := updateTfFiles(path, file.Name())
		filesUpdated = append(filesUpdated, filesEdited...)
	}
	slog.Debug("ci :: command :: files", "files", filesUpdated)
	return filesUpdated
}

func init() {
	cobra.OnInitialize(initConfig)
	checkForUpdatesCmd.AddCommand(ciCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ciCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}
