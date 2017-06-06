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

// This file contains types useful to load and manipulate docker files.

import (
	"io/ioutil"
	"regexp"
	"strings"
)

// Instruction represents each of the instructions that form an
// Dockerfile.
//
type DockerfileInstruction struct {
	// The name of the instruction.
	Name string

	// The arguments of the instruction.
	Args string
}

// Dockerfile conatins the information extracted from o docker file.
//
type Dockerfile struct {
	instructions []*DockerfileInstruction
}

// NewDockerfile creates a new empty docker file.
//
func NewDockerfile() *Dockerfile {
	return new(Dockerfile)
}

// Load loads a docker file and builds a list of structures containing
// the instruction names and arguments.
//
func (d *Dockerfile) Load(path string) *Dockerfile {
	d.instructions = make([]*DockerfileInstruction, 0)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	text := string(bytes)
	text = continueRe.ReplaceAllString(text, "")
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = commentRe.ReplaceAllString(line, "")
		if line == "" {
			continue
		}
		groups := FindRegexpGroups(line, instructionRe)
		if groups == nil {
			continue
		}
		name := groups["name"]
		args := groups["args"]
		instruction := &DockerfileInstruction{
			Name: name,
			Args: args,
		}
		d.instructions = append(d.instructions, instruction)
	}
	return d
}

// Instruction finds the first instruction within the given Dockerfile
// that haves the given name. If there is no such instruction then it
// returns nil.
//
func (d *Dockerfile) Instruction(name string) *DockerfileInstruction {
	for _, instruction := range d.instructions {
		if instruction.Name == name {
			return instruction
		}
	}
	return nil
}

// From finds the first FROM instruction within a Dockerfile. If that
// instruction exists then it returns its arguments. If it doesn't exist
// then it returns an empty string.
//
func (d *Dockerfile) From() string {
	from := d.Instruction("FROM")
	if from == nil {
		return ""
	}
	return from.Args
}

// Regular expressions used to process docker files.
//
var (
	commentRe     = regexp.MustCompile("\\s*#.*$")
	continueRe    = regexp.MustCompile("\\\\\n")
	instructionRe = regexp.MustCompile("^\\s*(?P<name>[a-zA-Z]+)\\s+(?P<args>.*)$")
)
