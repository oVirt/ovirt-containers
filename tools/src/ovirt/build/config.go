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
// the build configuration file.

import (
	"fmt"
	"os"

	"gopkg.in/ini.v1"
)

// BuildConfig contains the build configuration.
//
type BuildConfig struct {
	file  *ini.File
	build *ini.Section
}

// Version returns the version of the project stored in the
// configuration.
//
func (b *BuildConfig) Version() string {
	return b.build.Key("version").MustString("")
}

// Version returns the image prefix stored in the configuration.
//
func (b *BuildConfig) Prefix() string {
	return b.build.Key("prefix").MustString("")
}

// Images returns the path of the directory containing the source files
// of the image specifications.
//
func (b *BuildConfig) Images() string {
	return b.build.Key("images").MustString("")
}

// Registry returns the address of the Docker registry where images
// should be pushed to.
//
func (b *BuildConfig) Registry() string {
	return b.build.Key("registry").MustString("")
}

// Default global configuration data. This will be loaded first, and
// then the build.conf file will be loaded on top of it, overriding any
// value that is present.
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
var globalConfig *BuildConfig

// GlobalConfig returns the global configuration loaded from the
// build.conf file.
//
func GlobalConfig() *BuildConfig {
	if globalConfig == nil {
		var err error
		globalConfig = new(BuildConfig)
		globalConfig.file = ini.Empty()
		err = globalConfig.file.Append([]byte(globalData))
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Can't load default global configuration\n",
			)
			os.Exit(1)
		}
		err = globalConfig.file.Append("build.conf")
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Can't load global configuration file 'build.conf'\n",
			)
			os.Exit(1)
		}
		globalConfig.build = globalConfig.file.Section("build")
		if globalConfig.build == nil {
			fmt.Fprintf(
				os.Stderr,
				"The global configuration file doesn't contain a 'build' section\n",
			)
			os.Exit(1)
		}
	}
	return globalConfig
}
