package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Environment contains environment variable changes.
type Environment map[string]EnvValue

// EnvValue distinguishes between an empty value and a variable that must be removed.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads environment variable definitions from dir.
func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error while processing directory %s: %w", dir, err)
	}

	result := make(Environment, len(entries))
	for _, entry := range entries {
		if entry.Type().IsDir() {
			continue
		}
		if strings.Contains(entry.Name(), "=") {
			return nil, fmt.Errorf("file %s contains '=' in filename", entry.Name())
		}

		filePath := filepath.Join(dir, entry.Name())
		// The directory path is intentionally supplied by the envdir user.
		file, err := os.Open(filePath) //nolint:gosec
		if err != nil {
			return nil, fmt.Errorf("can't open a file %s: %w", filePath, err)
		}
		name, envValue, err := getValueFromFile(file)
		if err != nil {
			return nil, err
		}
		result[name] = *envValue
	}
	return result, nil
}

func getValueFromFile(file *os.File) (string, *EnvValue, error) {
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", nil, fmt.Errorf("can't get file info for %s: %w", file.Name(), err)
	}

	if info.Size() == 0 {
		return info.Name(), &EnvValue{NeedRemove: true}, nil
	}
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		firstLine := strings.ReplaceAll(scanner.Text(), "\x00", "\n")
		firstLine = strings.TrimRight(firstLine, " \t")
		return info.Name(), &EnvValue{Value: firstLine}, nil
	}
	if err := scanner.Err(); err != nil {
		return "", nil, fmt.Errorf("can't read from file %s: %w", file.Name(), err)
	}
	return "", nil, fmt.Errorf("in file={%s} value is not exist", info.Name())
}
