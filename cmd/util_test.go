package cmd

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/thundersparkf/samwise/cmd/errorHandlers"
	"testing"
)

func TestHappyCheckOutputFormat(t *testing.T) {
	csvLowerCaseTest, err := checkOutputFormat("csv")
	assert.Equal(t, "csv", csvLowerCaseTest)
	assert.Equal(t, nil, err)
	csvRandomCaseTest, err := checkOutputFormat("csv")
	assert.Equal(t, "csv", csvRandomCaseTest)
	assert.Equal(t, nil, err)
	jsonLowerCaseTest, err := checkOutputFormat("json")
	assert.Equal(t, "json", jsonLowerCaseTest)
	assert.Equal(t, nil, err)
	jsonRandomCaseTest, err := checkOutputFormat("json")
	assert.Equal(t, "json", jsonRandomCaseTest)
	assert.Equal(t, nil, err)

}

func TestUnhappyCheckOutputFormat(t *testing.T) {
	incorrectFormatTest, err := checkOutputFormat("testing")
	assert.Error(t, errors.New(errorHandlers.CheckOutputFormatError), err)
	assert.Equal(t, incorrectFormatTest, "")
}

func TestHappyCreateCSVReportFileNoData(t *testing.T) {
	data := []map[string]string{
		{"repo_link": "github.com/test_repo", "current_version": "2.4.4", "updates_available": "2.7.7|2.7.8"},
		{"repo_link": "github.com/test_repo_1", "current_version": "3.2.1", "updates_available": "3.2.2|3.2.3"},
	}
	createCSVReportFile(data, ".")
	results := readCsvFile("." + "/module_dependency_report.csv")
	fmt.Println(results)
	assert.Equal(t, len(results), 3)
	assert.Equal(t, results[1][0], data[0]["repo_link"])
	assert.Equal(t, results[1][1], data[0]["current_version"])
	assert.Equal(t, results[1][2], data[0]["updates_available"])

}

func TestUnhappyCreateCSVReportFileNoData(t *testing.T) {
	var data = make([]map[string]string, 0)
	createCSVReportFile(data, ".")
	results := readCsvFile("." + "/module_dependency_report.csv")
	assert.Equal(t, len(results), 1)

}

// TODO: Add test case to ensure only non-empty "updates_available" values get written to report
func TestUnhappyCreateCSVReportFileNilData(t *testing.T) {
	createCSVReportFile(nil, ".")
	results := readCsvFile("." + "/module_dependency_report.csv")
	assert.Equal(t, len(results), 1)

}
