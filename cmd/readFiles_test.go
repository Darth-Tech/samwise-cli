package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadDirectoryWithSlash(t *testing.T) {
	assert.Equal(t, "./test_dir", fixTrailingSlashForPath("./test_dir/"), "slash removed at the end")
}

func TestReadDirectoryWithoutSlash(t *testing.T) {
	assert.Equal(t, "./test_dir", fixTrailingSlashForPath("./test_dir"), "no slash at the end")
}
