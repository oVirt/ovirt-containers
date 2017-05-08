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
	file  *ini.File
	build *ini.Section
}

// Version returns the version of the project stored in the
// configuration.
//
func (p *ProjectConfig) Version() string {
	return p.build.Key("version").MustString("")
}

// Version returns the image prefix stored in the configuration.
//
func (p *ProjectConfig) Prefix() string {
	return p.build.Key("prefix").MustString("")
}

// Images returns the path of the directory containing the source files
// of the image specifications.
//
func (p *ProjectConfig) Images() string {
	return p.build.Key("images").MustString("")
}

// Registry returns the address of the Docker registry where images
// should be pushed to.
//
func (p *ProjectConfig) Registry() string {
	return p.build.Key("registry").MustString("")
}

// Default global project configuration data. This will be loaded first,
// and then the project.conf file will be loaded on top of it,
// overriding any value that is present.
//
const globalData = `
[build]
version=master
prefix=ovirt
images=image-specifications
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
		globalConfig.build = globalConfig.file.Section("build")
		if globalConfig.build == nil {
			fmt.Fprintf(
				os.Stderr,
				"The global configuration file '%s' doesn't contain a 'build' section\n",
				globalPath,
			)
			os.Exit(1)
		}
	}
	return globalConfig
}
