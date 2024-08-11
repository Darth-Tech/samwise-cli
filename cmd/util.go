package cmd

import (
	"encoding/csv"
	"errors"
	"github.com/thundersparkf/samwise/cmd/errorHandlers"
	"github.com/thundersparkf/samwise/cmd/outputs"
	"log/slog"
	"os"
	"slices"
	"strings"
)

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

func createCSVReportFile(data []map[string]string, path string) {
	slog.Debug("creating " + path + "/module_dependency_report.csv file")
	reportFilePath := path + "/module_dependency_report.csv"
	report, err := os.Create(reportFilePath)
	Check(err, "unable to create file ", reportFilePath)
	defer report.Close()

	writer := csv.NewWriter(report)
	defer writer.Flush()
	headers := []string{"repo_link", "current_version", "updates_available"}

	err = writer.Write(headers)
	Check(err, "unable to write headers to file", reportFilePath)
	for _, row := range data {
		if len(row["updates_available"]) > 0 {
			err := writer.Write([]string{row["repo_link"], row["current_version"], row["updates_available"]})
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

func readCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	Check(err, "util :: readCSVFile :: unable to read input file", filePath)
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	Check(err, "util :: readCSVFile :: unable to parse file as CSV", filePath)

	return records
}
