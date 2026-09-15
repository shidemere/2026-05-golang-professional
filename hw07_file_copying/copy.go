// Package main contains code for running program.
package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cheggaaa/pb/v3" //nolint:depguard
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	//nolint:gosec
	fromFile, err := os.Open(fromPath)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("souce file is not exist: %w", err)
	} else if err != nil {
		return fmt.Errorf("can't process input file: %w", err)
	}
	defer fromFile.Close()
	info, err := fromFile.Stat()
	if err != nil {
		return fmt.Errorf("can't extract from file metadata: %w", err)
	}
	isRegular := info.Mode().IsRegular()
	if !isRegular {
		return ErrUnsupportedFile
	}
	fromSize := info.Size()
	if offset != 0 {
		if offset > fromSize {
			return ErrOffsetExceedsFileSize
		}
		_, err := fromFile.Seek(offset, io.SeekStart)
		if err != nil {
			return fmt.Errorf("invalid offset: %w", err)
		}
	}
	//nolint:gosec
	toFile, err := os.OpenFile(toPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o666)
	if err != nil {
		return fmt.Errorf("can't create/write in file: %w", err)
	}
	defer toFile.Close()

	return copyWithProgress(toFile, fromFile, fromSize, offset, limit)
}

func copyWithProgress(toFile, fromFile *os.File, fromSize, offset, limit int64) error {
	var totalCopied int64
	chunkSize := 1024
	if limit == 0 {
		limit = fromSize - offset
	}
	if limit < int64(chunkSize) {
		chunkSize = int(limit)
	}

	bar := pb.StartNew(int(fromSize) - int(offset))
	for {
		n, err := io.CopyN(toFile, fromFile, int64(chunkSize))
		bar.Add(int(n))
		if errors.Is(err, io.EOF) {
			log.Println("copying completed successfully")
			break
		}
		if err != nil {
			return fmt.Errorf("something went wrong while copying: %w", err)
		}
		totalCopied += n
		remaining := limit - totalCopied
		if remaining < int64(chunkSize) {
			chunkSize = int(remaining)
		}
		if totalCopied == limit {
			break
		}
	}
	bar.Finish()
	return nil
}
