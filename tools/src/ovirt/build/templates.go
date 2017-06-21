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

// This file contains functions intended to process templates.

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// Context contains the objects passed to templates when they are
// evaluated.
//
type Context struct {
	project *Project
}

// ProcessTemplates scans all the files in the input directory,
// processes them as templates, and writes the result to the output
// directory.
//
func ProcessTemplates(project *Project, inDir string, outDir string) error {
	return filepath.Walk(inDir, func(inPath string, info os.FileInfo, err error) error {
		path, err := filepath.Rel(inDir, inPath)
		if err != nil {
			return err
		}
		outPath := filepath.Join(outDir, path)
		if info.IsDir() {
			return os.MkdirAll(outPath, info.Mode())
		}
		return ProcessTemplate(project, inPath, outPath)
	})
}

// ProcessTemplate loads the input file, processes it as a template, and
// writes the result to the output file.
//
func ProcessTemplate(project *Project, in string, out string) error {
	// Prepare the context object used for template evaluation:
	ctx := new(Context)
	ctx.project = project

	// Create the template and register the functions:
	tmpl := template.New(filepath.Base(in))
	tmpl.Funcs(template.FuncMap{
		"tag": func(name string) (string, error) {
			return tagFunc(ctx, name)
		},
	})

	// Parse the template:
	fmt.Printf(
		"Processing template '%s' and writing result to '%s'.\n",
		in,
		out,
	)
	_, err := tmpl.ParseFiles(in)
	if err != nil {
		return err
	}

	// Create the file where the output of the template evaluation
	// will be written, and remember to close it:
	info, err := os.Stat(in)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}
	defer file.Close()

	// Evaluate the template:
	return tmpl.Execute(file, ctx)
}

// tagFunc is a function intended to simplify writing templates that
// need to use the complete tag of an image, including the address of
// the registry. For example, a template for an OpenShift manifest that
// needs to specify the name of an image can be written as follows:
//
//	image: {{ tag "engine" }}
//
// With the default configuration that will be translated to this:
//
//	image: ovirt/engine:master
//
func tagFunc(context *Context, name string) (tag string, err error) {
	image, present := context.project.Images().Index()[name]
	if !present {
		err = fmt.Errorf("Can't find image for name '%s'", name)
		return
	}
	tag = image.Tag()
	return
}
