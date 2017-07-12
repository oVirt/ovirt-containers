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

// This is used to embed the project configuration into the binary of the tool,
// so that during run-time there is no need to have both the binary and the
// configuration files.
//
//go:generate go run scripts/embed.go -directory ../../.. -output tools/src/ovc/embedded.go project.conf image-specifications os-manifests

package main

// This tool loads and builds all the images.

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"ovc/build"
	"ovc/log"
)

// ToolFunc is the type of functions that implement tools.
//
type ToolFunc func(project *build.Project) error

// This index contains the mapping from names to tool functions.
//
var tools = map[string]ToolFunc{
	"build":  buildTool,
	"clean":  cleanTool,
	"deploy": deployTool,
	"push":   pushTool,
	"save":   saveTool,
}

// The name of the project file.
//
const conf = "project.conf"

func main() {
	// Get the name of the tool:
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s TOOL [ARGS...]\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	name := os.Args[1]

	// Find the function that corresponds to the tool name:
	tool := tools[name]
	if tool == nil {
		fmt.Fprintf(os.Stderr, "Can't find tool named '%s'.\n", name)
		os.Exit(1)
	}

	// Run the tool inside a different function, so that we can take
	// advantage of the 'defer' mechanism:
	os.Exit(run(name, tool))
}

func run(name string, tool ToolFunc) int {
	// Open the log:
	log.Open(name)
	log.Info("Log file is '%s'", log.Path())
	defer log.Close()

	// Check if the project file exists. If doesn't exist then we
	// need extract it, together with the rest of the source files
	// of the project, from the embedded data.
	file, _ := filepath.Abs(conf)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.Info("Extracting project")
		tmp, err := ioutil.TempDir("", "project")
		if err != nil {
			log.Error("Can't create temporary directory for project: %s", err)
			os.Exit(1)
		}
		defer os.RemoveAll(tmp)
		err = extractData(embedded, tmp)
		if err != nil {
			log.Error("Can't extract project: %s", err)
			os.Exit(1)
		}
		file = filepath.Join(tmp, conf)
	}

	// Load the project:
	log.Info("Loading project file '%s'", file)
	project, err := build.LoadProject(file)
	if err != nil {
		log.Error("%s", err)
		return 1
	}
	defer project.Close()

	// Call the tool function:
	log.Debug("Running tool '%s'", name)
	err = tool(project)
	if err != nil {
		log.Error("%s", err)
		log.Error("Tool failed, check log file '%s' for details", log.Path())
		return 1
	} else {
		log.Info("Tool finished successfully")
		return 0
	}
}

func extractData(data []byte, dir string) error {
	// Open the data archive:
	buffer := bytes.NewReader(data)
	expand, err := gzip.NewReader(buffer)
	if err != nil {
		return err
	}
	archive := tar.NewReader(expand)

	// Iterate through the entries of the archive and extract them
	// to the output directory:
	for {
		header, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeReg:
			err = extractFile(archive, header, dir)
		case tar.TypeDir:
			err = extractDir(archive, header, dir)
		default:
			err = nil
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func extractFile(archive *tar.Reader, header *tar.Header, dir string) error {
	// Create the file:
	path := filepath.Join(dir, header.Name)
	info := header.FileInfo()
	log.Debug("Extracting file '%s' to '%s'", header.Name, path)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy the contents:
	_, err = io.Copy(file, archive)
	return err
}

func extractDir(archive *tar.Reader, header *tar.Header, dir string) error {
	path := filepath.Join(dir, header.Name)
	info := header.FileInfo()
	log.Debug("Extracting directory '%s' to '%s'", header.Name, path)
	return os.Mkdir(path, info.Mode().Perm())
}
