package database

import (
	"compress/gzip"
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

func gzipCompress(inputFile, outputFile *os.File) error {
	if _, err := inputFile.Seek(0, 0); err != nil {
		return err
	}

	gzipWriter := gzip.NewWriter(outputFile)

	_, err := io.Copy(gzipWriter, inputFile)
	if err != nil {
		return err
	}

	if err := gzipWriter.Close(); err != nil {
		return err
	}

	return nil
}

func gzipDecompress(compressedFile, outputFile *os.File) error {
	gzipReader, err := gzip.NewReader(compressedFile)
	if err != nil {
		return err
	}

	const chunkSize = 10 * 1024 * 1024 // 10 MB
	for {
		if _, err := io.CopyN(outputFile, gzipReader, chunkSize); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	if err := gzipReader.Close(); err != nil {
		return err
	}

	return nil
}
