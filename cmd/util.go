package cmd

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"github.com/thundersparkf/samwise/cmd/errorHandlers"
	"github.com/thundersparkf/samwise/cmd/outputs"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type reportJson struct {
	Report []jsonReport `json:"report"`
}
type jsonReport struct {
	RepoLink         string `json:"repo_Link"`
	CurrentVersion   string `json:"current_version"`
	UpdatesAvailable string `json:"updates_available"`
	Error            string `json:"error"`
}

func Check(err error, message string, args ...any) {
	if err != nil {

		slog.Error(message, slog.Any("errorArgs", args))
		panic(err)
	}
}

func CheckNonPanic(err error, message string, args ...any) bool {
	if err != nil {
		slog.Error(message, slog.Any("errorArgs", args))
		slog.Error(err.Error())
		return true
	}
	return false
}
func generateReport(data []map[string]string, outputFilename string, outputFormat string, path string) {
	if outputFormat == outputs.CSV {
		createCSVReportFile(data, path, outputFilename)
	} else if outputFormat == outputs.JSON {
		createJSONReportFile(data, path, outputFilename)
	} else {
		Check(errors.New("output format "+outputFormat+"not available"), "")
	}

}

func createCSVReportFile(data []map[string]string, path string, filename string) {
	slog.Debug("creating " + path + "/" + filename + ".csv file")
	reportFilePath := path + "/" + filename + ".csv"
	report, err := os.Create(reportFilePath)
	Check(err, "util :: createCSVReportFile :: unable to create file ", reportFilePath)
	defer func(report *os.File) {
		err := report.Close()
		if err != nil {
			Check(err, "util :: createCSVReportFile :: unable to close file")
		}
	}(report)

	writer := csv.NewWriter(report)
	defer writer.Flush()
	headers := []string{"repo_link", "current_version", "updates_available"}

	err = writer.Write(headers)
	Check(err, "unable to write headers to file", reportFilePath)
	for _, row := range data {
		if len(row["updates_available"]) > 0 {
			err := writer.Write([]string{row["repo"], row["current_version"], row["updates_available"]})
			Check(err, "util :: CreateCSVReportFile :: unable to write record to file", row["repo"], row["current_version"], row["updates_available"])
			writer.Flush()
		}
	}
	slog.Debug("created " + reportFilePath)

}

func checkOutputFormat(outputFormat string) (string, error) {
	outputFormat = strings.ToLower(outputFormat)
	outputsAvailable := []string{outputs.CSV, outputs.JSON}
	if !slices.Contains(outputsAvailable, outputFormat) {
		return "", errors.New(errorHandlers.CheckOutputFormatError)
	} else {
		return strings.ToLower(outputFormat), nil
	}
}

func checkOutputFilename(outputFilename string) string {
	extension := filepath.Ext(outputFilename)
	outputFilename = strings.ReplaceAll(outputFilename, extension, "")
	return outputFilename

}

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

func createJSONReportFile(data []map[string]string, path string, filename string) {
	reportFilePath := path + "/" + filename + ".json"
	report, err := os.Create(reportFilePath)
	Check(err, "unable to create file ", reportFilePath)
	defer func(report *os.File) {
		err := report.Close()
		if err != nil {
			Check(err, "util :: createJSONReportFile :: unable to close file")
		}
	}(report)
	reportString, err := json.Marshal(data)
	Check(err, "util :: createJSONReportFile :: unable to marshal modules data")
	var reportJsonObject []jsonReport
	err = json.Unmarshal(reportString, &reportJsonObject)
	slog.Debug("util :: createJSONReportFile :: reportString :: " + string(reportString))
	Check(err, "util :: createJSONReportFile :: unable unmarshal into output format")
	var finalReportMap map[string][]jsonReport
	finalReportMap = map[string][]jsonReport{"report": reportJsonObject}
	reportOutputString, err := json.Marshal(finalReportMap)

	_, err = report.Write(reportOutputString)
	Check(err, "util :: createJSONReportFile :: unable to write to file", reportFilePath)
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
