package util

import (
	"encoding/csv"
	"fmt"
	"os"
)

func OutputWriter(input [][]string, fullFilePath string) error {
	file := &os.File{}
	if _, err := os.Stat(fullFilePath); os.IsNotExist(err) {
		file, err = os.Create(fullFilePath)
		if err != nil {
			return fmt.Errorf("could not create file: %v", err)
		}
	} else {
		file, err = os.OpenFile(fullFilePath, os.O_RDWR|os.O_APPEND, 0660)
		if err != nil {
			return fmt.Errorf("could not open file: %v", err)
		}
		// If file exists already we remove the header fields
		input = input[1:]
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range input {
		if err := writer.Write(value); err != nil {
			return fmt.Errorf("error writing value %s to csv: %v", input, err)
		}
	}

	return nil
}
