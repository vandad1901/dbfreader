package csv

import (
	"fmt"
	"os"
	"strings"

	"github.com/vandad1901/dbfreader/types"
)

func WriteToFile(data *types.DBFFile, fileName string) error {
	csvData, err := GetCSVData(data)
	if err != nil {
		return err
	}

	if err := os.WriteFile(fileName, csvData, 0o777); err != nil {
		return err
	}

	return nil
}

func GetCSVData(data *types.DBFFile) ([]byte, error) {
	var sb strings.Builder

	for i, ad := range data.ArrayDescriptors {
		if i > 0 {
			if _, err := sb.Write([]byte(",")); err != nil {
				return nil, fmt.Errorf("error writing to csv buffer: %w", err)
			}
		}
		if _, err := sb.Write([]byte(ad.FieldName)); err != nil {
			return nil, fmt.Errorf("error writing to csv buffer: %w", err)
		}
	}

	for _, record := range data.Records {
		if _, err := sb.Write([]byte("\n")); err != nil {
			return nil, fmt.Errorf("error writing to csv buffer: %w", err)
		}

		for i, ad := range data.ArrayDescriptors {
			if i > 0 {
				if _, err := sb.Write([]byte(",")); err != nil {
					return nil, fmt.Errorf("error writing to csv buffer: %w", err)
				}
			}
			if _, err := sb.Write([]byte(record[ad.FieldName].ToString())); err != nil {
				return nil, fmt.Errorf("error writing to csv buffer: %w", err)
			}
		}
	}

	return []byte(sb.String()), nil
}
