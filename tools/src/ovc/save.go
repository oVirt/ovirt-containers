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

// This tool saves the images to tar files.

import (
	"fmt"

	"ovc/build"
	"ovc/log"
)

func saveTool(project *build.Project) error {
	for _, image := range project.Images().List() {
		log.Info("Saving image '%s'", image)
		err := image.Save()
		if err != nil {
			return fmt.Errorf("Failed to save image '%s': %s", image, err)
		}
	}

	return nil
}
