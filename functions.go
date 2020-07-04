package main

import (
	"archive/tar"
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
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
	log.Printf("Creating tar...")
	var buffer bytes.Buffer
	compress(src, &buffer)

	fileToWrite, err := os.OpenFile(dest, os.O_CREATE|os.O_RDWR, os.FileMode(600))
	if err != nil {
		return err
	}

	if _, err := io.Copy(fileToWrite, &buffer); err != nil {
		return err
	}

	log.Printf("Done !")
	return nil
}

// Compress ... Compresses a directory into a tar writter buffer
func compress(src string, buf io.Writer) error {
	tw := tar.NewWriter(buf)
	sourcePath := filepath.ToSlash(src)

	filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(file)

		if sourcePath == header.Name {
			return nil
		}

		i := strings.Index(header.Name, "/")
		if i != -1 {
			header.Name = header.Name[i+1:]
		}

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

// RemoveEncyptionFromID ... Removes the encryption type from the id string
func RemoveEncyptionFromID(id string) string {
	idx := strings.Index(id, ":")
	if idx > -1 {
		return id[idx+1:]
	}

	return id
}

// SendEventMessage ... Sends a message to the rabbitMQ event queue
func SendEventMessage(eventMessage EventMessage) {
	log.Printf("%+v\n", eventMessage)
}
