package util

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"

	"github.com/sirupsen/logrus"
)

// StructToCsvOutput takes any struct and writes it to csv format
func StructToCsvOutput(stru interface{}, filename string) error {
	s := reflect.ValueOf(stru)
	typeOfI := s.Type()

	x := make([][]string, 2)
	for i := 0; i < s.NumField(); i++ {
		// Headers first
		x[0] = append(x[0], typeOfI.Field(i).Name)

		// Then content
		f := s.Field(i)
		content := fmt.Sprintf("%v", f.Interface())
		x[1] = append(x[1], content)
	}
	if err := outputWriter(x, filename); err != nil {
		logrus.Warningf("could not write instance health: %v", err)
	}

	return nil
}

func outputWriter(input [][]string, fullFilePath string) error {
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
