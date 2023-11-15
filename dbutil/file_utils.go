package dbutil

import (
	"compress/gzip"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
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
	if _, err := inputFile.Seek(0, 0); err != nil {
		return err
	}

	csvReader := csv.NewReader(inputFile)
	gzipWriter := gzip.NewWriter(outputFile)

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

	if err := gzipWriter.Close(); err != nil {
		return err
	}

	return nil
}

func gzipDecompress(inputFilePath, outputFilePath string) error {

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	inputFilePathFull := filepath.Clean(filepath.Join(currentDir, inputFilePath))
	outputFilePathFull := filepath.Clean(filepath.Join(currentDir, outputFilePath))

	gzipFile, err := os.Open(inputFilePathFull)
	if err != nil {
		return err
	}

	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return err
	}

	outputFile, err := os.Create(outputFilePathFull)
	if err != nil {
		return err
	}

	for {
		_, err := io.CopyN(outputFile, gzipReader, 1048576)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	if err := gzipReader.Close(); err != nil {
		return err
	}

	if err := gzipFile.Close(); err != nil {
		return err
	}

	if err := outputFile.Close(); err != nil {
		return err
	}

	return nil
}
