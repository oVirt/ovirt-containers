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

// This file contains types and functions used to load and manipulate
// the project configuration file.

import (
	"container/list"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/go-ini/ini"
)

// Project contains the project configuration.
//
type Project struct {
	work      string
	root      string
	version   string
	images    *ProjectImages
	manifests *ProjectManifests
}

// ProjectImages contains the information about the images that are part
// of the project.
//
type ProjectImages struct {
	project  *Project
	path     string
	prefix   string
	registry string
	list     []*Image
	index    map[string]*Image
}

// ProjectManifests contains the information about the manifests that
// are part of the project.
//
type ProjectManifests struct {
	project *Project
	path    string
}

// WorkingDirectory returns the absolute path of the working directory
// of the project.
//
func (p *Project) WorkingDirectory() string {
	return p.work
}

// Directory returns the the absolute path of the root directory of the
// project.
//
func (p *Project) Directory() string {
	return p.root
}

// Version returns the version of the project.
//
func (p *Project) Version() string {
	return p.version
}

// Images returns the information about the images that are part of the
// project.
//
func (p *Project) Images() *ProjectImages {
	return p.images
}

// Manifests returns the information about he OpenShift manifests that
// are part of the project.
//
func (p *Project) Manifests() *ProjectManifests {
	return p.manifests
}

// WorkingDirectory returns the absolute path of the working directory for the
// images of the project.
//
func (pi *ProjectImages) WorkingDirectory() string {
	return filepath.Join(pi.project.work, pi.path)
}

// Directory returns the absolute path of the directory containing the source
// files of the image specifications.
//
func (pi *ProjectImages) Directory() string {
	return filepath.Join(pi.project.root, pi.path)
}

// Prefix returns the prefix that should be used to tag the images of
// the project.
//
func (pi *ProjectImages) Prefix() string {
	return pi.prefix
}

// Registry returns the address of the Docker registry where images
// should be pushed to.
//
func (pi *ProjectImages) Registry() string {
	return pi.registry
}

// List returns a slice containing the images that are part of the
// project, sorted in the right build order.
//
func (pi *ProjectImages) List() []*Image {
	return pi.list
}

// Index returns a map containing the images that are part of the
// project, indexed by name.
//
func (pi *ProjectImages) Index() map[string]*Image {
	return pi.index
}

// WorkingDirectory returns the absolute path of the working directory for the
// OpenShift manifests of the project.
//
func (pm *ProjectManifests) WorkingDirectory() string {
	return filepath.Join(pm.project.work, pm.path)
}

// Directory returns the absolute path of the directory containing the
// source files of the OpenShift manifests.
//
func (pm *ProjectManifests) Directory() string {
	return filepath.Join(pm.project.root, pm.path)
}

// Close releases all the resources used by the project, including the
// temporary directory used to store the results of processsing
// templates. Once the project is closed it can no longer be used.
//
func (p *Project) Close() error {
	return os.RemoveAll(p.work)
}

// Default project file name.
//
const defaultProjectFile = "project.conf"

// Default global project configuration data. This will be loaded first,
// and then the 'project.conf' file will be loaded on top of it,
// overriding any value that is present.
//
const defaultProjectData = `
version=master

[images]
prefix=ovirt
directory=image-specifications
registry=

[manifests]
directory=os-manifests
`

// LoadProject loads a project from the given path. If the path is empty
// then it will load the project from the 'project.conf' file inside the
// current working directory.
//
func LoadProject(path string) (project *Project, err error) {
	var file *ini.File
	var section *ini.Section

	// Create an initially empty project:
	project = new(Project)

	// Calculate the absolute path of the project:
	root, _ := filepath.Abs(filepath.Dir(path))
	project.root = root

	// Create a temporary directory that will be used to store the
	// results of generating files from templates, and maybe other
	// temporary files.
        project.work, err = ioutil.TempDir("", "work")
        if err != nil {
		err = fmt.Errorf("Can't create temporary work directory: %s\n", err)
		return
        }

	// If the path is empty then use the current directory and the
	// default project file name:
	if path == "" {
		path, err = os.Getwd()
		if err != nil {
			return
		}
		path = filepath.Join(path, defaultProjectFile)
	}

	// Load the default project data:
	file = ini.Empty()
	err = file.Append([]byte(defaultProjectData))
	if err != nil {
		return
	}

	// Load the project file:
	err = file.Append(path)
	if err != nil {
		return
	}

	// Copy the main parameters from the configuration to the
	// project object:
	section = file.Section("")
	project.version = section.Key("version").MustString("")

	// Load the images:
	err = loadImages(file, project)
	if err != nil {
		return
	}

	// Load the manifests:
	err = loadManifests(file, project)
	if err != nil {
		return
	}

	return
}

// Regular expression used to extract the name and version of an image
// from the value of the FROM instruction of a Dockerfile.
//
var fromRe = regexp.MustCompile("^(?P<prefix>.*)/(?P<name>[^/]+):(?P<version>[^:]+)$")

// loadImages images scans the images directory, loads the descriptions
// and stores them into the project.
//
// The images will have their parent resolved, but only if the parent is
// also in the same images directory, otherwise it will be nil.
//
func loadImages(file *ini.File, project *Project) error {
	// Check that the 'images' section is available:
	section := file.Section("images")
	if section == nil {
		return fmt.Errorf("The project configuration doesn't contain the 'images' section\n")
	}

	// Load basic attributes:
	images := new(ProjectImages)
	project.images = images
	images.project = project
	images.path = section.Key("directory").MustString("")
	images.prefix = section.Key("prefix").MustString("")
	images.registry = section.Key("registry").MustString("")

	// The source files of images may be templates, and those
	// templates may refer to some properties of other images. In
	// particular the Dockerfile of one image may refer to its
	// parent using a template:
	//
	//	FROM {{ tag "base" }}
	//
	// In order to support that we need first to do a basic loading
	// of the images, to have at least the name. After that we can
	// load the image details, which will process the templates.
	images.list = []*Image{}
	paths, err := filepath.Glob(filepath.Join(project.root, images.path, "*"))
	if err != nil {
		return err
	}
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			image := NewImage(project, filepath.Base(path))
			if err != nil {
				return err
			}
			images.list = append(images.list, image)
		}
	}
	images.index = make(map[string]*Image)
	for _, image := range images.list {
		images.index[image.name] = image
	}

	// Now that we have the names of all the images, we can process
	// load the details of the images.
	for _, image := range images.list {
		image.Load()
	}

	// Resolve the references to parent images, using the FROM
	// instruction:
	for _, image := range project.images.list {
		from := image.Dockerfile().From()
		if from == "" {
			continue
		}
		groups := FindRegexpGroups(from, fromRe)
		if groups == nil {
			continue
		}
		prefix := groups["prefix"]
		if prefix != project.Images().Prefix() {
			continue
		}
		name := groups["name"]
		parent := project.images.index[name]
		if parent != nil {
			image.parent = parent
		}
	}

	// Sort the loaded images in build order:
	sortImages(project.images.list)

	return nil
}

// loadImage loads the details of the image, including the content of
// the Dockerfile, if it exists.
//
func loadImage(image *Image) error {
	// Check if there is a Dockerfile in the directory, and if it
	// does then load it:
	path := filepath.Join(image.Directory(), "Dockerfile")
	if _, err := os.Stat(path); err == nil {
		image.dockerfile = NewDockerfile()
		image.dockerfile.Load(path)
	}

	return nil
}


// sortImages implements a topological sort of the images according to
// their dependencies. If image A is the base for image B, then image A
// is guaranteed to appear before imabe B in the sorted slice. this is
// intended to be able to build images in the right order.
//
func sortImages(images []*Image) {
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

// loadManifests images scans the OpenShift manifests directory, loads
// the descriptions and stores them into the project.
//
func loadManifests(file *ini.File, project *Project) error {
	// Check that the 'manifests' section is available:
	section := file.Section("manifests")
	if section == nil {
		return fmt.Errorf("The project configuration doesn't contain the 'manifests' section\n")
	}

	// Create the manifests object, and get the directory:
	manifests := new(ProjectManifests)
	project.manifests = manifests
	manifests.project = project
	manifests.path = section.Key("directory").MustString("")

	// Process the templates:
	return ProcessTemplates(
		project,
		manifests.Directory(),
		manifests.WorkingDirectory(),
	)
}
