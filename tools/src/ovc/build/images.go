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

package build

// This file contains types and function used to load and manipulate the
// descriptions of images.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Image contains the description of an image, as well as methods to
// build it.
//
type Image struct {
	project    *Project
	work       string
	directory  string
	name       string
	tag        string
	dockerfile *Dockerfile
	parent     *Image
}

// NewImage creates a new empty image with the given name and belonging
// to the given project.
//
func NewImage(project *Project, name string) *Image {
	i := new(Image)
	i.project = project
	i.name = name
	return i
}

// Load loads the details of the image, including the content of
// the Dockerfile, if it exists.
//
func (i *Image) Load() error {
	// Calculate the tag:
	i.tag = fmt.Sprintf(
		"%s/%s:%s",
		i.project.Images().Prefix(),
		i.name,
		i.project.Version(),
	)
	registry := i.project.Images().Registry()
	if registry != "" {
		i.tag = fmt.Sprintf(
			"%s/%s",
			registry,
			i.tag,
		)
	}

	// Calculate the directory:
	i.directory = filepath.Join(i.project.Images().Directory(), i.name)

	// Process the templates, if needed:
	if i.work == "" {
		i.work = filepath.Join(i.project.Images().WorkingDirectory(), i.name)
		err := ProcessTemplates(i.project, i.directory, i.work)
		if err != nil {
			return err
		}
	}

        // Check if there is a Dockerfile in the directory, and if it
        // does then load it:
        dockerfilePath := filepath.Join(i.work, "Dockerfile")
        if _, err := os.Stat(dockerfilePath); err == nil {
                i.dockerfile = NewDockerfile()
                i.dockerfile.Load(dockerfilePath)
        }

        return nil
}


// Directory returns the path of the directory that contains the source
// files of the image.
//
func (i *Image) Directory() string {
	return i.directory
}

// Name returns the name of the image.
//
func (i *Image) Name() string {
	return i.name
}

// Tag returns the tag of the image.
//
func (i *Image) Tag() string {
	return i.tag
}

// Dockerfile returns the object that describes the Dockerfile used by
// the image.
//
func (i *Image) Dockerfile() *Dockerfile {
	return i.dockerfile
}

// Parent returns the parent image.
//
func (i *Image) Parent() *Image {
	return i.parent
}

// String returns a string representation of the image.
//
func (i *Image) String() string {
	return i.Tag()
}

// Build builds the given image.
//
func (i *Image) Build() error {
	return RunCommand(
		"docker",
		"build",
		fmt.Sprintf("--tag=%s", i.Tag()),
		i.work,
	)
}

// Saves the image to a tar file.
//
func (i *Image) Save() error {
	// Get the tag of the image and calculate a file name from it,
	// replacing all characters that aren't convenient in file names
	// (slashes and colons) with dashes:
	tag := i.Tag()
	replacer := strings.NewReplacer(
		"/", "-",
		":", "-",
	)
	path := replacer.Replace(tag) + ".tar"

	// Save the image to a tar file:
	err := RunCommand(
		"docker",
		"save",
		tag,
		fmt.Sprintf("--output=%s", path),
	)
	if err != nil {
		return err
	}

	// Compress the tar file, using gzip:
	return RunCommand(
		"gzip",
		"--force",
		path,
	)
}

// Push pushes the image to the docker registry.
//
func (i *Image) Push() error {
	return RunCommand(
		"docker",
		"push",
		i.Tag(),
	)
}

// Remove removes the image from the local docker storage.
//
func (i *Image) Remove() error {
	return RunCommand(
		"docker",
		"rmi",
		i.Tag(),
	)
}