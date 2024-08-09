package cmd

import "testing"

func TestReadDirectoryWithSlash(t *testing.T) {
	if fixTrailingSlashForPath("./test_dir/") == "./test_dir" {
		t.Logf("Success")
	} else {
		t.Errorf("expected %s received %s", fixTrailingSlashForPath("./test_dir"), "./test_dir")
	}
}

func TestReadDirectoryWithoutSlash(t *testing.T) {
	if fixTrailingSlashForPath("./test_dir") == "./test_dir" {
		t.Logf("Success")
	} else {
		t.Errorf("expected %s received %s", fixTrailingSlashForPath("./test_dir"), "./test_dir")
	}
}
