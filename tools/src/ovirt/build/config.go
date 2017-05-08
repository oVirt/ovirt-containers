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
	"fmt"
	"os"

	"gopkg.in/ini.v1"
)

// ProjectConfig contains the project configuration.
//
type ProjectConfig struct {
	file   *ini.File
	images *ImagesConfig
}

// ImagesConfig contains the configuration items related to images.
//
type ImagesConfig struct {
	section *ini.Section
}

// Images returns the set of items related to images.
//
func (p *ProjectConfig) Images() *ImagesConfig {
	if p.images == nil {
		section := p.file.Section("images")
		if section != nil {
			p.images = new(ImagesConfig)
			p.images.section = section
		}
	}
	return p.images
}

// Version returns the version that should be used to tag the images of
// the project.
//
func (i *ImagesConfig) Version() string {
	return i.section.Key("version").MustString("")
}

// Version returns the prefix that should be used to tag the images of
// the project.
//
func (i *ImagesConfig) Prefix() string {
	return i.section.Key("prefix").MustString("")
}

// Directory returns the path of the directory containing the source
// files of the image specifications.
//
func (i *ImagesConfig) Directory() string {
	return i.section.Key("directory").MustString("")
}

// Registry returns the address of the Docker registry where images
// should be pushed to.
//
func (i *ImagesConfig) Registry() string {
	return i.section.Key("registry").MustString("")
}

// Default global project configuration data. This will be loaded first,
// and then the project.conf file will be loaded on top of it,
// overriding any value that is present.
//
const globalData = `
[images]
version=master
prefix=ovirt
directory=image-specifications
registry=localhost:5000
`

// The global configuration, which will be lazily loaded the first time
// tha the GlobalConfig function is called.
//
var globalConfig *ProjectConfig

// GlobalConfig returns the global configuration loaded from the
// project.conf file.
//
func GlobalConfig() *ProjectConfig {
	if globalConfig == nil {
		var err error
		globalConfig = new(ProjectConfig)
		globalConfig.file = ini.Empty()
		err = globalConfig.file.Append([]byte(globalData))
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Can't load default global configuration\n",
			)
			os.Exit(1)
		}
		globalPath := "project.conf"
		err = globalConfig.file.Append(globalPath)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Can't load global configuration file '%s'\n",
				globalPath,
			)
			os.Exit(1)
		}
		images := globalConfig.Images()
		if images == nil {
			fmt.Fprintf(
				os.Stderr,
				"The global configuration file '%s' doesn't contain a 'images' section\n",
				globalPath,
			)
			os.Exit(1)
		}
	}
	return globalConfig
}
