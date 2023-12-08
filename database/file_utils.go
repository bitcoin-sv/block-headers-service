package database

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"

	gz "github.com/klauspost/compress/gzip"
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

func gzipFastCompress(inputFile, outputFile *os.File) error {
	if _, err := inputFile.Seek(0, 0); err != nil {
		return err
	}

	gzipWriter, err := gz.NewWriterLevel(outputFile, gz.BestSpeed)
	if err != nil {
		return fmt.Errorf("creating gzip writer: %w", err)
	}
	defer func() {
		if closeErr := gzipWriter.Close(); closeErr != nil {
			fmt.Printf("gzipWriter close error: %v", closeErr)
		}
	}()

	if _, err := io.Copy(gzipWriter, inputFile); err != nil && err != io.EOF {
		return fmt.Errorf("copying content to gzip writer: %w", err)
	}

	return nil
}

func gzipCompress(inputFile, outputFile *os.File) error {
	if _, err := inputFile.Seek(0, 0); err != nil {
		return err
	}

	gzipWriter, err := gzip.NewWriterLevel(outputFile, gzip.BestSpeed)
	if err != nil {
		return fmt.Errorf("creating gzip writer: %w", err)
	}
	defer func() {
		if closeErr := gzipWriter.Close(); closeErr != nil {
			fmt.Printf("gzipWriter close error: %v", closeErr)
		}
	}()

	if _, err := io.Copy(gzipWriter, inputFile); err != nil && err != io.EOF {
		return fmt.Errorf("copying content to gzip writer: %w", err)
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

func gzipDecompressWithBuffer(compressedFile, outputFile *os.File) error {
	gzipReader, err := gz.NewReader(compressedFile)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := gzipReader.Close(); closeErr != nil {
			fmt.Printf("gzipReader close error: %v", closeErr)
		}
	}()

	bufferSize := 16 * 1024 * 1024 // 16 MB buffer size
	buffer := make([]byte, bufferSize)

	// io.CopyBuffer for efficient decompression
	if _, err = io.CopyBuffer(outputFile, gzipReader, buffer); err != nil && err != io.EOF {
		return err
	}

	return nil
}
