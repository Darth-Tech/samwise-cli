package cmd

import (
	"encoding/csv"
	"errors"
	"github.com/thundersparkf/samwise/cmd/errorHandlers"
	"github.com/thundersparkf/samwise/cmd/outputs"
	"log"
	"log/slog"
	"os"
	"slices"
	"strings"
)

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func CreateCSVReportFile(data []map[string]string, path string) {
	slog.Debug("creating " + path + "/module_dependency_report.csv file")
	file2, err := os.Create(path + "/module_dependency_report.csv")
	if err != nil {
		panic(err)
	}
	defer file2.Close()

	writer := csv.NewWriter(file2)
	defer writer.Flush()
	// this defines the header value and data values for the new csv file
	headers := []string{"repo_link", "current_version", "updates_available"}

	err = writer.Write(headers)
	Check(err)
	for _, row := range data {
		err := writer.Write([]string{row["repo_link"], row["current_version"], row["updates_available"]})
		Check(err)
		writer.Flush()
	}
	slog.Debug("created " + path + "/module_dependency_report.csv file")

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
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return records
}
