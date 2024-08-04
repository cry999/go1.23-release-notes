package main

import (
	"archive/tar"
	"log"
	"os"
)

const filename = "archive.tar"

func main() {
	createTar()

	f, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}

	hdr, err := tar.FileInfoHeader(fileInfoNames{FileInfo: f}, "readme.txt")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Uname: %q, Gname: %q", hdr.Uname, hdr.Gname)
}

type fileInfoNames struct{ os.FileInfo }

func (fin fileInfoNames) Uname() string { return "uname" }
func (fin fileInfoNames) Gname() string { return "gname" }

func createTar() {
	os.Remove(filename)

	dist, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer dist.Close()

	tw := tar.NewWriter(dist)
	defer tw.Close()

	files := []struct {
		Name, Body string
	}{
		{"readme.txt", "This archive contains some text files."},
		{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
		{"todo.txt", "Get animal handling license."},
	}
	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0o600,
			Size: int64(len(file.Body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			log.Fatal(err)
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			log.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		log.Fatal(err)
	}
}
