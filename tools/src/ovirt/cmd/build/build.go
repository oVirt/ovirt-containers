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

	"ovirt/build"
)

func main() {
	// Load the project:
	project, err := build.LoadProject("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't load project: %s\n", err);
		os.Exit(1)
	}
	defer project.Close()

	// Build the images:
	for _, image := range project.Images().List() {
		fmt.Printf("Building image '%s'.\n", image)
		err := image.Build()
		if err != nil {
			fmt.Fprintf(
				os.Stderr, "Failed to build image '%s'.\n",
				image,
			)
			os.Exit(1)
		}
	}
}
