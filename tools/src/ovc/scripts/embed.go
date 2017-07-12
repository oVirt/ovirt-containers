/*
Copyright (c) 2017 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

// This file is intended to be used as an script. It takes the files
// given in the command line, creates a compresses tarball, and
// generates a Go source file that contains a data variable with the
// contents of that tarball. For example, to following command line will
// generate an embedded.go file that contains the the project.conf file
// and os-manifests directory:
//
//	go run embed.go -output embedded.go project.conf os-manifests
//
// The generated embedded.go file can then be added to the project, and
// used to extract the embedded data during run-time.

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// Option for the directory where the script should change to before
	// doing anything else:
	var directoryFlag string
	flag.StringVar(
		&directoryFlag,
		"directory",
		"",
		"change to `directory` before doing anything else",
	)

	// Option for the name of the output file:
	var outputFlag string
	flag.StringVar(
		&outputFlag,
		"output",
		"embedded.go",
		"name of the output `file`",
	)

	// Prepare the usage message:
	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"Usage: %s [OPTIONS] FILE ...\n",
			filepath.Base(os.Args[0]),
		)
		flag.PrintDefaults()
	}

	// Parse the command line:
	flag.Parse()

	// Check that there is at least one file to add:
	paths := flag.Args()
	if len(paths) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// Change directory:
	if directoryFlag != "" {
		err := os.Chdir(directoryFlag)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Can't change to directory '%s': %s\n",
				directoryFlag,
				err,
			)
			os.Exit(1)
		}
	}

	// Create a tar archive writer that will compress and write the
	// result to an in-memory buffer:
	buffer := new(bytes.Buffer)
	compress := gzip.NewWriter(buffer)
	archive := tar.NewWriter(compress)

	// Add paths to the tar archive:
	for _, path := range paths {
		addTree(archive, path)
	}

	// Close the tarball and the gzip stream:
	if err := archive.Close(); err != nil {
		log.Fatalln(err)
	}
	if err := compress.Close(); err != nil {
		log.Fatalln(err)
	}

	// Create the output file:
	outputFile, err := os.Create(outputFlag)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Can't create output file '%s': %s\n",
			outputFlag,
			err,
		)
		os.Exit(1)
	}

	// Generate the Go source code that contains the compressed
	// tarball:
	fmt.Fprintf(outputFile, "package main\n")
	fmt.Fprintf(outputFile, "\n")
	fmt.Fprintf(outputFile, "var embedded = []byte{\n")
	for _, datum := range buffer.Bytes() {
		fmt.Fprintf(outputFile, "\t0x%02x,\n", datum)
	}
	fmt.Fprintf(outputFile, "}\n")

	// Close the output file:
	outputFile.Close()
}

func addTree(archive *tar.Writer, base string) error {
	return filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		// Stop inmediately if something fails when trying to
		// walk the directory:
		if err != nil {
			return err
		}

		// Add an entry to the tarball:
		switch {
		case info.Mode().IsRegular():
			return addFile(archive, path, info)
		case info.Mode().IsDir():
			return addDir(archive, path, info)
		default:
			return nil
		}
	})
}

func addFile(archive *tar.Writer, path string, info os.FileInfo) error {
	// Add the header:
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = path
	err = archive.WriteHeader(header)
	if err != nil {
		return err
	}

	// Add the content:
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(archive, file)
	return err
}

func addDir(archive *tar.Writer, path string, info os.FileInfo) error {
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = path + "/"
	return archive.WriteHeader(header)
}
