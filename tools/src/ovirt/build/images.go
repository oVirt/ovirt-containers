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
	"container/list"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Image contains the description of an image, as well as methods to
// build it.
//
type Image struct {
	path       string
	name       string
	version    string
	tag        string
	dockerfile *Dockerfile
	parent     *Image
}

// NewImage creates a new image object, initially empty.
//
func NewImage() *Image {
	return new(Image)
}

// Load Load an image specification from the the directory with the
// given path.
//
func (i *Image) Load(path string) *Image {
	// Get the global configuration:
	config := GlobalConfig()

	// Allocate the image, using the name of the directory as its
	// name:
	i.path = path
	i.name = filepath.Base(path)

	// Check if there is a Dockerfile in the directory, and if it
	// does then load it:
	dockerfilePath := filepath.Join(path, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err == nil {
		dockerfile := NewDockerfile()
		dockerfile.Load(dockerfilePath)
		i.dockerfile = dockerfile
	}

	// Set the version:
	i.version = config.Images().Version()

	return i
}

// Path returns the path of the directory that contains the source files
// of the image.
//
func (i *Image) Path() string {
	return i.path
}

// Name returns the name of the image.
//
func (i *Image) Name() string {
	return i.name
}

// Version returns the version of the image.
//
func (i *Image) Version() string {
	return i.version
}

// Tag returns the tag of the image. If no tag has been explicitly
// assigned then a default one will be calculated based on the
// configuratoin and the image name.
//
func (i *Image) Tag() string {
	// If the tag has a value, then return it:
	if i.tag != "" {
		return i.tag
	}

	// Otherwise build a default tag, using the prefix taken from
	// the build configuration, the image name, and the image
	// version:
	config := GlobalConfig()
	return fmt.Sprintf("%s/%s:%s", config.Images().Prefix(), i.name, i.version)
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
		i.Path(),
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
	// Get the global configuration:
	config := GlobalConfig()

	// Tag the image adding the registry name as a prefix, as this
	// is needed in order to push it:
	imageTag := i.Tag()
	registryTag := fmt.Sprintf("%s/%s", config.Images().Registry(), imageTag)
	err := RunCommand(
		"docker",
		"tag",
		imageTag,
		registryTag,
	)
	if err != nil {
		return err
	}

	// Push the image:
	return RunCommand(
		"docker",
		"push",
		registryTag,
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

// Regular expression used to extract the name and version of an image
// from the value of the FROM instruction of a Dockerfile.
//
var fromRe = regexp.MustCompile("^(?P<prefix>.*)/(?P<name>[^/]+):(?P<version>[^:]+)$")

// LoadImages images scans the images directory, loads the descriptions
// and returns them sorted in build order.
//
// The images will have their parent resolved, but only if the parent is
// also in the same images directory, otherwise it will be nil.
//
func LoadImages() []*Image {
	// Get the global configuration:
	config := GlobalConfig()

	// Walk all the files under the top level images directory and
	// find all the diretories that contain a Dockerfile file. Those
	// are the directories that will be considered as image
	// directories.
	images := []*Image{}
	filepath.Walk(config.Images().Directory(), func(path string, info os.FileInfo, err error) error {
		if info.Name() == "Dockerfile" {
			image := NewImage().Load(filepath.Dir(path))
			images = append(images, image)
		}
		return err
	})

	// Resolve the references to parent images, using the FROM
	// instruction:
	index := make(map[string]*Image)
	for _, image := range images {
		index[image.Name()] = image
	}
	for _, image := range index {
		from := image.Dockerfile().From()
		if from == "" {
			continue
		}
		groups := FindRegexpGroups(from, fromRe)
		if groups == nil {
			continue
		}
		prefix := groups["prefix"]
		if prefix != config.Images().Prefix() {
			continue
		}
		name := groups["name"]
		parent := index[name]
		if parent != nil {
			image.parent = parent
		}
	}

	// Sor the loaded images in build order:
	SortImages(images)

	return images
}

// SortImages implements a topological sort of the images according to
// their dependencies. If image A is the base for image B, then image A
// is guaranteed to appear before imabe B in the sorted slice. this is
// intended to be able to build images in the right order.
//
func SortImages(images []*Image) {
	index := 0
	visited := make(map[*Image]bool)
	visited[nil] = true
	pending := list.New()
	for _, image := range images {
		visited[image] = false
		pending.PushBack(image)
	}
	for pending.Len() > 0 {
		current := pending.Remove(pending.Front()).(*Image)
		parent := current.Parent()
		if visited[parent] {
			visited[current] = true
			images[index] = current
			index++
		} else {
			pending.PushBack(parent)
			pending.PushBack(current)
		}
	}
}
