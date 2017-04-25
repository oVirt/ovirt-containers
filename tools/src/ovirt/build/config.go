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
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Config provides mechanims to load and query a set of configuration
// parameters.
//
// This configuration can be loaded from a file with a syntax similar to
// java properties files, or to the /etc/sysconfig/files used by Linux
// distributions like Fedora and CentOS.
//
type Config struct {
	values map[string]string
}

// NewConfig creates a new configuratoin object.
//
func NewConfig() *Config {
	c := new(Config)
	c.values = make(map[string]string)
	return c
}

// Load loads the configuration from given reader.
//
func (c *Config) Load(reader io.Reader) error {
	// Read line by line, discarding empty lines and comments, and
	// jonining continuation lines. The result will be stored in a
	// temporary slice that will be later be processed to extract
	// the actual values.
	lines := make([]string, 0)
	scanner := bufio.NewScanner(reader)
	buffer := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if line[0] == '#' || line[0] == ';' {
			continue
		}
		last := len(line) - 1
		if line[last] == '\\' {
			buffer += line[:last]
			buffer += " "
		} else {
			buffer += line
			lines = append(lines, buffer)
			buffer = ""
		}
	}
	if buffer != "" {
		lines = append(lines, buffer)
	}

	// Process the lines and extract the values:
	for _, line := range lines {
		index := strings.Index(line, "=")
		if index != -1 {
			name := strings.TrimSpace(line[:index])
			value := strings.TrimSpace(line[index+1:])
			c.values[name] = value
		}
	}

	return nil
}

// LoadFile loads the configuration from the file with the given path.
//
func (c *Config) LoadFile(path string) error {
	// Create a reader to read the string:
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Load the file:
	return c.Load(file)
}

// LoadText loads the configuration from the given text.
//
func (c *Config) LoadText(text string) error {
	return c.Load(strings.NewReader(text))
}

// Has checks if the configuration contains the given parameter.
//
func (c *Config) Has(name string) bool {
	_, has := c.values[name]
	return has
}

// Get returns the string containing the value of the given parameter.
//
func (c *Config) Get(name string) string {
	value, has := c.values[name]
	if has {
		return value
	}
	return ""
}

// BuildConfig contains the build configuration.
//
type BuildConfig struct {
	config *Config
}

// NewBuildConfig creates a new empty build configuration.
//
func NewBuildConfig() *BuildConfig {
	b := new(BuildConfig)
	b.config = NewConfig()
	return b
}

// LoadFile loads the build configuratoin from the given file.
//
func (b *BuildConfig) LoadFile(path string) error {
	return b.config.LoadFile(path)
}

// LoadText loads the build configuratoin from the given text.
//
func (b *BuildConfig) LoadText(path string) error {
	return b.config.LoadText(path)
}

// Version returns the version of the project stored in the
// configuration.
//
func (b *BuildConfig) Version() string {
	return b.config.Get("version")
}

// Version returns the image prefix stored in the configuration.
//
func (b *BuildConfig) Prefix() string {
	return b.config.Get("prefix")
}

// Images returns the path of the directory containing the source files
// of the image specifications.
//
func (b *BuildConfig) Images() string {
	return b.config.Get("images")
}

// Default global configuration data. This will be loaded first, and
// then the build.conf file will be loaded on top of it, overriding any
// value that is present.
//
const globalData = `
version=master
prefix=ovirt
images=image-specifications
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
		globalConfig = NewBuildConfig()
		err = globalConfig.LoadText(globalData)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Can't load default global configuration\n",
			)
			os.Exit(1)
		}
		err = globalConfig.LoadFile("build.conf")
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Can't load global configuration file 'build.conf'\n",
			)
			os.Exit(1)
		}
	}
	return globalConfig
}
