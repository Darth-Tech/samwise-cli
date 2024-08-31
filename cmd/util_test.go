package cmd

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thundersparkf/samwise/cmd/errorHandlers"
)

// Functions to help testing
func readCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	Check(err, "util :: readCSVFile :: unable to read input file", filePath)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			Check(err, "util :: readCsvFile :: unable to close file")
		}
	}(f)

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	Check(err, "util :: readCSVFile :: unable to parse file as CSV", filePath)

	return records
}

func readJSONFile(filePath string) reportJson {
	var report reportJson
	file, err := os.Open(filePath)
	Check(err, "util :: readJSONFile :: unable to open file", filePath)
	byteValue, err := io.ReadAll(file)
	Check(err, "util :: readJSONFile :: unable to read bytes", byteValue)
	err = json.Unmarshal(byteValue, &report)
	Check(err, "util :: readJSONFile :: unable to unmarshal json", string(byteValue))
	return report
}

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

func TestCheckOutputFilename(t *testing.T) {
	assert.Equal(t, "./module", checkOutputFilename("./module.csv"))
	assert.Equal(t, "./module_test", checkOutputFilename("./module_test.json"))
	assert.Equal(t, "./module", checkOutputFilename("./module"))

	assert.Equal(t, "module_test", checkOutputFilename("module_test.csv"))
	assert.Equal(t, "", checkOutputFilename(".pdf"))
}
func TestGenerateReport(t *testing.T) {
	data := []map[string]string{
		{"repo": "github.com/test_repo", "current_version": "2.4.4", "updates_available": "2.7.7|2.7.8", "file_name": "main.tf"},
		{"repo": "github.com/test_repo_1", "current_version": "3.2.1", "updates_available": "3.2.2|3.2.3", "file_name": "test/main.tf"},
	}
	generateReport(data, "module_dependency_report", "csv", ".")
	resultsCSV := readCsvFile("." + "/module_dependency_report.csv")
	assert.Equal(t, len(resultsCSV), 3, "csv report unable to generated")
	generateReport(data, "module_dependency", "json", ".")
	resultsJSON := readJSONFile("./module_dependency.json")
	assert.Equal(t, len(resultsJSON.Report), 2, "json report unable to generated")
	defer func() {
		if r := recover(); r != nil {
			assert.PanicsWithValue(t, "output format yaml not available", func() { generateReport(data, "module_dependency", "yaml", ".") }, "not panicking when incorrect output format is given")
		}
	}()
}

func TestGenerateFailureReport(t *testing.T) {
	data := []map[string]string{
		{"repo": "github.com/test_repo", "current_version": "2.4.4", "updates_available": "2.7.7|2.7.8", "file_name": "main.tf", "error": "random error"},
		{"repo": "github.com/test_repo_1", "current_version": "3.2.1", "updates_available": "3.2.2|3.2.3", "file_name": "main.tf"},
	}
	createJSONReportFile(data, ".", "failure_report")
	results := readJSONFile("./failure_report.json")
	assert.Equal(t, len(results.Report), 2)
	assert.Equal(t, results.Report[0].RepoLink, data[0]["repo"], "repo_link key is not matching")
	assert.Equal(t, results.Report[0].CurrentVersion, data[0]["current_version"], "current_version key is not matching")
	assert.Equal(t, results.Report[0].UpdatesAvailable, data[0]["updates_available"], "updates_available key is not matching")
	assert.Equal(t, results.Report[0].FileName, data[0]["file_name"], "file_name is not matching")
	assert.Equal(t, results.Report[0].Error, data[0]["error"], "error key is not matching")
	assert.Equal(t, results.Report[1].Error, data[1]["error"], "error key is not matching")

}
func TestHappyCreateCSVReportFile(t *testing.T) {
	data := []map[string]string{
		{"repo": "github.com/test_repo", "current_version": "2.4.4", "updates_available": "2.7.7|2.7.8", "file_name": "main.tf"},
		{"repo": "github.com/test_repo_1", "current_version": "3.2.1", "updates_available": "3.2.2|3.2.3", "file_name": "main.tf"},
	}
	createCSVReportFile(data, ".", "module_report")
	results := readCsvFile("." + "/module_report.csv")
	assert.Equal(t, len(results), 3)
	assert.Equal(t, data[0]["repo"], results[1][0], "repo link mismatch")
	assert.Equal(t, data[0]["current_version"], results[1][1], "current_version mismatch")
	assert.Equal(t, data[0]["updates_available"], results[1][3], "updates_available mismatch")
	assert.Equal(t, data[0]["file_name"], results[1][2], "file_name mismatch")

}

func TestHappyCreateCSVReportFileLatestVersion(t *testing.T) {
	data := []map[string]string{
		{"repo": "github.com/test_repo", "current_version": "2.4.4", "updates_available": "2.7.7|2.7.8", "latest_version": "2.7.8", "file_name": "main.tf"},
		{"repo": "github.com/test_repo_1", "current_version": "3.2.1", "updates_available": "3.2.2|3.2.3", "latest_version": "3.2.3", "file_name": "test/main.tf"},
	}
	// Set cobra flag for testing
	createCSVReportFile(data, ".", "module_report")
	results := readCsvFile("." + "/module_report.csv")
	assert.Equal(t, len(results), 3)
	assert.Equal(t, data[0]["repo"], results[1][0], "repo link mismatch")
	assert.Equal(t, data[0]["current_version"], results[1][1], "current_version mismatch")
	assert.Equal(t, data[0]["updates_available"], results[1][3], "latest_version mismatch")
	assert.Equal(t, data[0]["file_name"], results[1][2], "file_name mismatch")

}

func TestUnhappyCreateCSVReportFileNoData(t *testing.T) {
	var data = make([]map[string]string, 0)
	createCSVReportFile(data, ".", "module_dependency_report")
	results := readCsvFile("." + "/module_dependency_report.csv")
	assert.Equal(t, len(results), 1)

}

// TODO: Add test case to ensure only non-empty "updates_available" values get written to report
func TestUnhappyCreateCSVReportFileNilData(t *testing.T) {
	createCSVReportFile(nil, ".", "module_dependency_report")
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
		{"repo": "github.com/test_repo", "current_version": "2.4.4", "updates_available": "2.7.7|2.7.8", "file_name": "main.tf"},
		{"repo": "github.com/test_repo_1", "current_version": "3.2.1", "updates_available": "3.2.2|3.2.3", "file_name": "test/main.tf"},
	}
	createJSONReportFile(data, ".", "module_dependency")
	results := readJSONFile("." + "/module_dependency.json")
	assert.Equal(t, len(results.Report), 2)
	assert.Equal(t, results.Report[0].RepoLink, data[0]["repo"], "repo_link key is not matching")
	assert.Equal(t, results.Report[0].CurrentVersion, data[0]["current_version"], "current_version key is not matching")
	assert.Equal(t, results.Report[0].UpdatesAvailable, data[0]["updates_available"], "updates_available key is not matching")
	assert.Equal(t, results.Report[0].FileName, data[0]["file_name"], "file_name key is not matching")

}

func TestUnhappyCreateJSONReportFileNoData(t *testing.T) {
	var data = make([]map[string]string, 0)
	var expectedReport = reportJson{[]jsonReport(nil)}
	createJSONReportFile(data, ".", "module_dependency_report")
	results := readJSONFile("." + "/module_dependency_report.json")
	assert.Equal(t, expectedReport, results, "report not empty")
	assert.Empty(t, results.Report, "reports are non-zero")
}

func TestReadTfFiles(t *testing.T) {

	// TESTS TO READ FILE AND CHECK MODULE SOURCES

}

func TestGetGreatestSemverFromList(t *testing.T) {
	list1 := "1.0.0|1.0.1|1.0.5|1.0.3-beta|1.0.3-alpha"
	list2 := "1.0.0|v1.0.1|v1.0.3-beta|v1.0.5-alpha"
	list3 := "1.0.0|1.0.1|1.0.3-beta|1.0.5-alpha|v1.0.5-beta"
	list4 := "1.0.dwd0|1.0wvwv.1|1.0.3-beta|1.0.5-alpha|v1.0.5-beta"

	assert.Equal(t, "1.0.5", getGreatestSemverFromList(list1))
	assert.Equal(t, "v1.0.5-alpha", getGreatestSemverFromList(list2))
	assert.Equal(t, "v1.0.5-beta", getGreatestSemverFromList(list3))
	assert.Equal(t, "v1.0.5-beta", getGreatestSemverFromList(list4))

}
