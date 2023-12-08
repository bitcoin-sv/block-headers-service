package database

import (
	"encoding/csv"
	"os"
	"testing"
	"time"
)

// to run this bench test run
// go test -bench=. -count 5 -benchtime=10s -run=^#
// or
// go test -bench=. -count 2 -run=^#

func BenchmarkGzipFastCompress(b *testing.B) {
	compressedInputFile, err := os.CreateTemp("", "benchmark_input.csv")
	if err != nil {
		b.Fatalf("Error creating input file: %v", err)
	}
	defer os.Remove(compressedInputFile.Name())

	writer := csv.NewWriter(compressedInputFile)
	defer writer.Flush()

	headers := []string{"hash", "version", "merkleroot", "nonce", "bits", "chainwork", "timestamp", "cumulatedWork"}
	if err := writer.Write(headers); err != nil {
		b.Fatalf("Error writing headers into input file: %v", err)
	}

	// Generate 100 million rows with dummy data
	for i := 0; i < 100000000; i++ {
		row := []string{"hash", "1", "merkle", "123", "456", "chain", "123456789", "987654321"}
		if err := writer.Write(row); err != nil {
			panic(err)
		}
	}

	// Create a temporary file for the compressed output
	outputFile, err := os.CreateTemp("", "benchmark_output.gz")
	if err != nil {
		b.Fatalf("Error creating temporary output file: %v", err)
	}
	defer os.Remove(outputFile.Name())

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := compressedInputFile.Seek(0, 0); err != nil {
			b.Fatalf("Error seeking input file: %v", err)
		}

		startTime := time.Now()
		err := gzipFastCompress(compressedInputFile, outputFile)
		elapsedTime := time.Since(startTime)

		if err != nil {
			b.Errorf("Compression error: %v", err)
		}

		if _, err := outputFile.Seek(0, 0); err != nil {
			b.Errorf("Error seeking output file: %v", err)
		}
		// Calculate and report the time taken per operation
		b.ReportMetric(float64(time.Second)/elapsedTime.Seconds(), "ops/s")
	}
}

func BenchmarkGzipCompress(b *testing.B) {
	compressedInputFile, err := os.CreateTemp("", "benchmark_input.csv")
	if err != nil {
		b.Fatalf("Error creating input file: %v", err)
	}
	defer os.Remove(compressedInputFile.Name())

	writer := csv.NewWriter(compressedInputFile)
	defer writer.Flush()

	headers := []string{"hash", "version", "merkleroot", "nonce", "bits", "chainwork", "timestamp", "cumulatedWork"}
	if err := writer.Write(headers); err != nil {
		b.Fatalf("Error writing headers into input file: %v", err)
	}

	// Generate 100 million rows with dummy data
	for i := 0; i < 100000000; i++ {
		row := []string{"hash", "1", "merkle", "123", "456", "chain", "123456789", "987654321"}
		if err := writer.Write(row); err != nil {
			panic(err)
		}
	}

	// Create a temporary file for the compressed output
	outputFile, err := os.CreateTemp("", "benchmark_output.gz")
	if err != nil {
		b.Fatalf("Error creating temporary output file: %v", err)
	}
	defer os.Remove(outputFile.Name())

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := compressedInputFile.Seek(0, 0); err != nil {
			b.Fatalf("Error seeking input file: %v", err)
		}

		startTime := time.Now()
		err := gzipCompress(compressedInputFile, outputFile)
		elapsedTime := time.Since(startTime)

		if err != nil {
			b.Errorf("Compression error: %v", err)
		}

		if _, err := outputFile.Seek(0, 0); err != nil {
			b.Errorf("Error seeking output file: %v", err)
		}
		// Calculate and report the time taken per operation
		b.ReportMetric(float64(time.Second)/elapsedTime.Seconds(), "ops/s")
	}
}

func BenchmarkGzipDecompressWithBuffer(b *testing.B) {
	// Generate a temporary sample input file
	sampleInputFile, err := os.CreateTemp("", "benchmark_input.csv")
	if err != nil {
		b.Fatalf("Error creating sample input file: %v", err)
	}
	defer os.Remove(sampleInputFile.Name())

	writer := csv.NewWriter(sampleInputFile)
	defer writer.Flush()

	headers := []string{"hash", "version", "merkleroot", "nonce", "bits", "chainwork", "timestamp", "cumulatedWork"}
	if err := writer.Write(headers); err != nil {
		b.Fatalf("Error writing headers into sample input file: %v", err)
	}

	// Generate 100 million rows with dummy data
	for i := 0; i < 100000000; i++ {
		row := []string{"hash", "1", "merkle", "123", "456", "chain", "123456789", "987654321"}
		if err := writer.Write(row); err != nil {
			b.Fatalf("Error writing row into sample input file: %v", err)
		}
	}

	// Create a temporary file for the compressed input
	compressedInputFile, err := os.CreateTemp("", "benchmark_input.gz")
	if err != nil {
		b.Fatalf("Error creating compressed input file: %v", err)
	}
	defer os.Remove(compressedInputFile.Name())

	// Compress the sample input file
	err = gzipFastCompress(sampleInputFile, compressedInputFile)
	if err != nil {
		b.Fatalf("Error compressing input file: %v", err)
	}

	// Create an output file for decompression
	outputFile, err := os.CreateTemp("", "benchmark_output.csv")
	if err != nil {
		b.Fatalf("Error creating output file: %v", err)
	}
	defer os.Remove(outputFile.Name())

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset the input file position
		if _, err := compressedInputFile.Seek(0, 0); err != nil {
			b.Fatalf("Error seeking input file: %v", err)
		}

		// Reset the output file position
		if _, err := outputFile.Seek(0, 0); err != nil {
			b.Fatalf("Error seeking output file: %v", err)
		}

		startTime := time.Now()
		err := gzipDecompressWithBuffer(compressedInputFile, outputFile)
		elapsedTime := time.Since(startTime)

		if err != nil {
			b.Errorf("Decompression error: %v", err)
		}

		// Calculate and report the time taken per operation
		b.ReportMetric(float64(time.Second)/elapsedTime.Seconds(), "ops/s")
	}
}

func BenchmarkGzipDecompress(b *testing.B) {
	// Generate a temporary sample input file
	sampleInputFile, err := os.CreateTemp("", "benchmark_input.csv")
	if err != nil {
		b.Fatalf("Error creating sample input file: %v", err)
	}
	defer os.Remove(sampleInputFile.Name())

	writer := csv.NewWriter(sampleInputFile)
	defer writer.Flush()

	headers := []string{"hash", "version", "merkleroot", "nonce", "bits", "chainwork", "timestamp", "cumulatedWork"}
	if err := writer.Write(headers); err != nil {
		b.Fatalf("Error writing headers into sample input file: %v", err)
	}

	// Generate 100 million rows with dummy data
	for i := 0; i < 100000000; i++ {
		row := []string{"hash", "1", "merkle", "123", "456", "chain", "123456789", "987654321"}
		if err := writer.Write(row); err != nil {
			b.Fatalf("Error writing row into sample input file: %v", err)
		}
	}

	// Create a temporary file for the compressed input
	compressedInputFile, err := os.CreateTemp("", "benchmark_input.gz")
	if err != nil {
		b.Fatalf("Error creating compressed input file: %v", err)
	}
	defer os.Remove(compressedInputFile.Name())

	// Compress the sample input file
	err = gzipFastCompress(sampleInputFile, compressedInputFile)
	if err != nil {
		b.Fatalf("Error compressing input file: %v", err)
	}

	// Create an output file for decompression
	outputFile, err := os.CreateTemp("", "benchmark_output.csv")
	if err != nil {
		b.Fatalf("Error creating output file: %v", err)
	}
	defer os.Remove(outputFile.Name())

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset the input file position
		if _, err := compressedInputFile.Seek(0, 0); err != nil {
			b.Fatalf("Error seeking input file: %v", err)
		}

		// Reset the output file position
		if _, err := outputFile.Seek(0, 0); err != nil {
			b.Fatalf("Error seeking output file: %v", err)
		}

		startTime := time.Now()
		err := gzipDecompress(compressedInputFile, outputFile)
		elapsedTime := time.Since(startTime)

		if err != nil {
			b.Errorf("Decompression error: %v", err)
		}

		// Calculate and report the time taken per operation
		b.ReportMetric(float64(time.Second)/elapsedTime.Seconds(), "ops/s")
	}
}
