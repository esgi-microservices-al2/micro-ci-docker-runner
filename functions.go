package main

import (
	"archive/tar"
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
)

// Getenv ... Retrives an environment variable but provides a default fallback value if empty
func Getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// FailOnError ... A simple function to handle errors
func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// CreateTar ... Creates a tar from a directory
func CreateTar(src string, dest string) error {
	var buffer bytes.Buffer
	compress(src, &buffer)

	fileToWrite, err := os.OpenFile(dest, os.O_CREATE|os.O_RDWR, os.FileMode(600))
	if err != nil {
		return err
	}

	if _, err := io.Copy(fileToWrite, &buffer); err != nil {
		return err
	}

	return nil
}

// Compress ... Compresses a directory into a tar writter buffer
func compress(src string, buf io.Writer) error {
	tw := tar.NewWriter(buf)

	filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(file)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})

	if err := tw.Close(); err != nil {
		return err
	}

	return nil
}
