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

func TestHappyCreateCSVReportFile(t *testing.T) {
	data := []map[string]string{
		{"repo": "github.com/test_repo", "current_version": "2.4.4", "updates_available": "2.7.7|2.7.8"},
		{"repo": "github.com/test_repo_1", "current_version": "3.2.1", "updates_available": "3.2.2|3.2.3"},
	}
	createCSVReportFile(data, ".")
	results := readCsvFile("." + "/module_dependency_report.csv")
	fmt.Println(results)
	assert.Equal(t, len(results), 3)
	assert.Equal(t, data[0]["repo"], results[1][0], "repo link mismatch")
	assert.Equal(t, data[0]["current_version"], results[1][1], "current_version mismatch")
	assert.Equal(t, data[0]["updates_available"], results[1][2], "updates_available mismatch")

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

func TestCheckError(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.NotPanics(t, nil, func() { Check(nil, "testing non panic case") })
			assert.PanicsWithValue(t, "test error", func() { Check(errors.New("test error"), "error testing") })
			assert.PanicsWithValue(t, "test error", func() { Check(errors.New("test error"), "error testing", "testArg1", "testArg2") })
		}
	}()

}

func TestCheckNonPanic(t *testing.T) {
	assert.Equal(t, true, CheckNonPanic(errors.New("non panic error triggered"), "testing triggering non panic error"))
	assert.Equal(t, false, CheckNonPanic(nil, "testing triggering non panic error"))
	assert.Equal(t, false, CheckNonPanic(nil, ""))

}

func TestHappyCreateJSONReportFileNoData(t *testing.T) {
	data := []map[string]string{
		{"repo_link": "github.com/test_repo", "current_version": "2.4.4", "updates_available": "2.7.7|2.7.8"},
		{"repo_link": "github.com/test_repo_1", "current_version": "3.2.1", "updates_available": "3.2.2|3.2.3"},
	}
	createJSONReportFile(data, ".")
	results := readJSONFile("." + "/module_dependency_report.json")
	assert.Equal(t, len(results.Report), 2)
	assert.Equal(t, results.Report[0].RepoLink, data[0]["repo_link"], "repo_link key is not matching")
	assert.Equal(t, results.Report[0].CurrentVersion, data[0]["current_version"], "current_version key is not matching")
	assert.Equal(t, results.Report[0].UpdatesAvailable, data[0]["updates_available"], "updates_available key is not matching")

}

func TestUnhappyCreateJSONReportFileNoData(t *testing.T) {
	var data = make([]map[string]string, 0)
	var expectedReport = reportJson{[]jsonReport{}}
	createJSONReportFile(data, ".")
	results := readJSONFile("." + "/module_dependency_report.json")
	assert.Equal(t, expectedReport, results, "report not empty")
	assert.Empty(t, results.Report, "reports are non-zero")
}
