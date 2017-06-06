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

// This tool loads and builds all the images.

import (
	"fmt"
	"os"
	"path/filepath"

	"ovirt/build"
)

type ToolFunc func(project *build.Project) error

var toolsIndex = map[string]ToolFunc{
	"build":  buildTool,
	"clean":  cleanTool,
	"deploy": deployTool,
	"push":   pushTool,
	"save":   saveTool,
}

func main() {
	// Get the name of the tool:
	if len(os.Args) < 2 {
		fmt.Fprintf(
			os.Stderr,
			"Usage: %s TOOL [ARGS...]\n",
			filepath.Base(os.Args[0]),
		)
		os.Exit(1)
	}
	toolName := os.Args[1]

	// Find the function that corresponds to the tool name:
	toolFunc := toolsIndex[toolName]
	if toolFunc == nil {
		fmt.Fprintf(
			os.Stderr,
			"Can't find tool named '%s'.\n",
			toolName,
		)
		os.Exit(1)
	}

	// Load the project:
	project, err := build.LoadProject("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't load project: %s\n", err)
		os.Exit(1)
	}

	// Call the tool function, and close the project regardless of
	// the result:
	err = toolFunc(project)

	// Check the result of the tool:
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	// Bye:
	os.Exit(0)
}
