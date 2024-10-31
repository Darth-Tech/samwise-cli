package cmd

import (
	"errors"
	"github.com/rs/zerolog/log"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirectorySearch(t *testing.T) {
	DirectoriesToIgnore = []string{"errorHandlers"}
	Depth = 1
	filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if path == "hcl_types/module_type.go" {
			isDir, skipDir := directorySearch(".", path, d)
			assert.Equal(t, isDir, false, path+" is not a directory but picked as one")
			assert.Empty(t, skipDir, "skipDir is not nil")
		} else if path == "hcl_types" {
			isDir, skipDir := directorySearch(".", path, d)
			assert.Equal(t, isDir, true, path+" is a directory but not picked as one")
			assert.Empty(t, skipDir, "skipDir is not nil")
		} else if path == "errorHandlers" {
			isDir, skipDir := directorySearch(".", path, d)
			assert.Equal(t, isDir, true, path+" is a directory but not picked as one")
			assert.Equal(t, skipDir, fs.SkipDir, "skipDir is not fs.SkipDir")
		}
		return nil
	})

}

func TestDirectorySearchInWalk(t *testing.T) {
	rootDir := "../.github"
	Depth = 0
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		Check(err, "checkForUpdates :: command :: ", path)
		isAllowedDir, dirError := directorySearch(rootDir, path, d)
		if errors.Is(dirError, fs.SkipDir) {
			return dirError
		}
		if strings.Contains(path, "workflows") {
			assert.Equal(t, true, isAllowedDir, path+"should be skipped due to depth of file walk but is not skipped")
		}
		return nil
	})
	if err != nil {
		log.Error().Msgf("error in walking directory %s", err.Error())
	}
}
