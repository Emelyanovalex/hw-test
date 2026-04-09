package main

import (
	"errors"
	"io"
	"os"

	pb "github.com/cheggaaa/pb/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	src, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}

	if !info.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	fileSize := info.Size()
	if offset > fileSize {
		return ErrOffsetExceedsFileSize
	}

	if _, err = src.Seek(offset, io.SeekStart); err != nil {
		return err
	}

	bytesToCopy := fileSize - offset
	if limit > 0 && limit < bytesToCopy {
		bytesToCopy = limit
	}

	dst, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	bar := pb.Full.Start64(bytesToCopy)
	defer bar.Finish()

	reader := bar.NewProxyReader(src)
	_, err = io.CopyN(dst, reader, bytesToCopy)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}
