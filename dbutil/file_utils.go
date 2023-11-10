package dbutil

import (
	"compress/gzip"
	"encoding/csv"
	"io"
	"os"
)

func fileExistsAndIsReadable(path string) bool {
	fileInfo, err := os.Stat(path)
	if err == nil {
		if fileInfo.Mode()&os.ModeType == 0 {
			return true
		}
	}
	return false
}

func createDirectory(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func gzipCompress(inputFile, outputFile *os.File) error {
	inputFile.Seek(0, 0)

	csvReader := csv.NewReader(inputFile)

	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()

	csvWriter := csv.NewWriter(gzipWriter)
	defer csvWriter.Flush()

	for {
		record, err := csvReader.Read()
		if err != nil {
			break // Reached the end of the CSV file
		}

		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func gzipDecompress(inputFilePath, outputFilePath string) error {
	gzipFile, err := os.Open(inputFilePath)
	if err != nil {
		return err
	}
	defer gzipFile.Close()

	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, gzipReader)
	if err != nil {
		return err
	}

	return nil
}
