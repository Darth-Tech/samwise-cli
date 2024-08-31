package cmd

import (
	"io/fs"
	"path/filepath"
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
