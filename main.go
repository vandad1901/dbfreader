package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/vandad1901/dbfreader/csv"
	"github.com/vandad1901/dbfreader/internal/bytes"
	"github.com/vandad1901/dbfreader/types"
)

func main() {
	iName := flag.String("i", "", "the input method, available options are: dbf")
	outputMode := flag.String("output", "csv", "the output method, available options are: csv, debug")
	oName := flag.String("o", "", "the input method, available options are: dbf")

	flag.Parse()

	if *iName == "" {
		fmt.Println("Please enter input file name")
		flag.Usage()
		os.Exit(1)
	}

	var (
		data *types.DBFFile
		err  error
	)

	data, err = bytes.ReadFromFile(*iName)
	if err != nil {
		logError(err)
		os.Exit(1)
	}

	switch *outputMode {
	case "debug":
		fmt.Printf("header: %+v\n\ndescriptors: %+v\n\nrecords: %+v\n", data.Header, data.ArrayDescriptors, data.Records)
	case "csv":
		if *oName == "" {
			fmt.Print("Please enter output file name\n\n")
			flag.Usage()
			os.Exit(1)
		}
		if err := csv.WriteToFile(data, *oName); err != nil {
			logError(err)
			os.Exit(1)
		}
	default:
		logError(errors.New("unexpected output mode"))
		os.Exit(1)
	}
}

func logError(e error) {
	_, err := fmt.Fprintf(os.Stderr, "%s\n", e)
	if err != nil {
		panic(err)
	}
}
