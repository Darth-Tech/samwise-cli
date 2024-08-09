package cmd

import (
	"bufio"
	"os"
)

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func CreateCSVReportFile(path string) (*os.File, *bufio.Writer) {
	f, _ := os.Create(path)
	writer := bufio.NewWriter(f)

	return f, writer
}
