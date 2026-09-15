package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCopy(t *testing.T) {
	tests := []struct {
		name     string
		offset   int64
		limit    int64
		expected string
	}{
		{
			name:     "whole file",
			expected: "out_offset0_limit0.txt",
		},
		{
			name:     "limit 10",
			limit:    10,
			expected: "out_offset0_limit10.txt",
		},
		{
			name:     "limit 1000",
			limit:    1000,
			expected: "out_offset0_limit1000.txt",
		},
		{
			name:     "limit exceeds file size",
			limit:    10000,
			expected: "out_offset0_limit10000.txt",
		},
		{
			name:     "offset 100 limit 1000",
			offset:   100,
			limit:    1000,
			expected: "out_offset100_limit1000.txt",
		},
		{
			name:     "offset 6000 limit 1000",
			offset:   6000,
			limit:    1000,
			expected: "out_offset6000_limit1000.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(t.TempDir(), "output.txt")
			if err := Copy(filepath.Join("testdata", "input.txt"), outputPath, tt.offset, tt.limit); err != nil {
				t.Fatalf("Copy() returned an error: %v", err)
			}

			actual, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("failed to read output file: %v", err)
			}
			expected, err := os.ReadFile(filepath.Join("testdata", tt.expected))
			if err != nil {
				t.Fatalf("failed to read expected file: %v", err)
			}

			if !bytes.Equal(actual, expected) {
				t.Errorf("copied content differs from %s: got %d bytes, want %d", tt.expected, len(actual), len(expected))
			}
		})
	}
}

func TestCopyOffsetExceedsFileSize(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "output.txt")
	err := Copy(filepath.Join("testdata", "input.txt"), outputPath, 10000, 0)
	if !errors.Is(err, ErrOffsetExceedsFileSize) {
		t.Fatalf("Copy() error = %v, want %v", err, ErrOffsetExceedsFileSize)
	}
}

func TestCopyUnsupportedFile(t *testing.T) {
	inputPath := t.TempDir()
	outputPath := filepath.Join(t.TempDir(), "output.txt")

	err := Copy(inputPath, outputPath, 0, 0)
	if !errors.Is(err, ErrUnsupportedFile) {
		t.Fatalf("Copy() error = %v, want %v", err, ErrUnsupportedFile)
	}
}

func TestCopySourceDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()
	err := Copy(filepath.Join(tempDir, "missing.txt"), filepath.Join(tempDir, "output.txt"), 0, 0)
	if err == nil {
		t.Fatal("Copy() error = nil, want an error")
	}
}

func TestCopyTruncatesDestination(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "output.txt")
	if err := os.WriteFile(outputPath, bytes.Repeat([]byte("x"), 100), 0o600); err != nil {
		t.Fatalf("failed to prepare output file: %v", err)
	}

	if err := Copy(filepath.Join("testdata", "input.txt"), outputPath, 0, 10); err != nil {
		t.Fatalf("Copy() returned an error: %v", err)
	}

	actual, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	expected, err := os.ReadFile(filepath.Join("testdata", "out_offset0_limit10.txt"))
	if err != nil {
		t.Fatalf("failed to read expected file: %v", err)
	}

	if !bytes.Equal(actual, expected) {
		t.Errorf("destination was not truncated correctly: got %d bytes, want %d", len(actual), len(expected))
	}
}
