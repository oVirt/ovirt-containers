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

// This tool deploys the application to the OpenShift cluster.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"ovc/build"
)

// The name of the project.
//
const (
	projectName  = "ovirt"
	projectTitle = "oVirt"
)

// Names of the required service accounts.
//
const (
	useRoot        = "useroot"
	privilegedUser = "privilegeduser"
)

func deployTool(project *build.Project) error {
	var err error

	// Check that the 'oc' tool is available and that it is the
	// right version:
	err = validateOc()
	if err != nil {
		return err
	}

	// Log in as system administrator:
	err = runOc(
		"login",
		"-u",
		"system:admin",
	)
	if err != nil {
		return err
	}

	// Create the project:
	err = runOc(
		"new-project",
		projectName,
		"--description="+projectTitle,
		"--display-name="+projectTitle,
	)
	if err != nil {
		return err
	}

	// Add administrator permissios to the 'developer' user account:
	err = runOc(
		"adm",
		"policy",
		"add-role-to-user",
		"admin",
		"developer",
		"-n",
		projectName,
	)
	if err != nil {
		return err
	}

	// Create a service account that can use the root user:
	err = runOc(
		"create",
		"serviceaccount",
		useRoot,
	)
	if err != nil {
		return err
	}
	err = runOc(
		"adm",
		"policy",
		"add-scc-to-user",
		"anyuid",
		"-z",
		useRoot,
	)
	if err != nil {
		return err
	}

	// Create a service account that can has access to advanced host
	// privileges, for use inside the VSDC pod:
	err = runOc(
		"create",
		"serviceaccount",
		privilegedUser,
	)
	if err != nil {
		return err
	}
	err = runOc(
		"adm",
		"policy",
		"add-scc-to-user",
		"privileged",
		"-z",
		privilegedUser,
	)
	if err != nil {
		return err
	}

	// Create engine and VDSC deployments and add them to the
	// project:
	err = runOc(
		"create",
		"-f",
		project.Manifests().WorkingDirectory(),
		"-R",
	)
	if err != nil {
		return err
	}

	// Change the host name for the engine deployment, according to
	// the host name that was assigned to the associated route, then
	// unpause it:
	engineHost, err := evalOc(
		"get",
		"routes",
		"ovirt-engine",
		"--output=jsonpath={.spec.host}",
	)
	if err != nil {
		return err
	}
	spiceProxyHost, err := evalOc(
		"get",
		"routes",
		"ovirt-spice-proxy",
		"--output=jsonpath={.spec.host}",
	)
	if err != nil {
		return err
	}
	err = runOc(
		"set",
		"env",
		"dc/ovirt-engine",
		"-c",
		"ovirt-engine",
		"OVIRT_FQDN="+engineHost,
		"SPICE_PROXY=http://"+spiceProxyHost+":3128",
	)
	if err != nil {
		return err
	}
	err = runOc(
		"patch",
		"dc/ovirt-engine",
		"--patch",
		`{
			"spec": {
				"paused": false
			}
		}`,
	)
	if err != nil {
		return err
	}

	return nil
}

// Regular expression used to extract the version of the 'oc' tool from
// the first line of output of the 'oc version' command. The typical
// output from that command is something like this:
//
//	oc v1.5.0+031cbe4
//	kubernetes v1.5.2+43a9be4
//	features: Basic-Auth GSSAPI Kerberos SPNEGO
//
// This regular expression captures the major, minor and micro version
// numbers of the first line.
//
var ocVersionRe = regexp.MustCompile(
	"^oc\\s+v(?P<major>\\d+)\\.(?P<minor>\\d+)\\.(?P<micro>\\d+).*$",
)

// Minimum supported 'oc' version.
//
const (
	minOcMajor = 1
	minOcMinor = 5
)

// ValidateOc checks that the OpenShift 'oc' tool is installed and that
// it is the right version. If the validation fails it prints an error
// message and aborts the application.
//
func validateOc() error {
	// Get the value of the PATH environment variable:
	path, present := os.LookupEnv("PATH")
	if !present {
		return fmt.Errorf(
			"The PATH environment variable isn't set, can't locate the 'oc' tool'",
		)
	}

	// Check that the 'oc' tools is available in one of the
	// directories specified in the PATH environment variable:
	dirs := strings.Split(path, string(os.PathListSeparator))
	exec := ""
	for _, dir := range dirs {
		file := filepath.Join(dir, "oc")
		info, err := os.Lstat(file)
		if err != nil || info.IsDir() {
			continue
		}
		if info.Mode()|0111 != 0 {
			exec = file
			break
		}
	}
	if exec == "" {
		return fmt.Errorf(
			"Can't find the 'oc' tool in the path",
		)
	}

	// Run the tool to extract the version number, and check that it
	// is what we expect:
	out := build.EvalCommand("oc", "version")
	if out == nil {
		return fmt.Errorf("Failed to run 'oc version'")
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) < 1 {
		return fmt.Errorf("Output of 'oc version' is empty")
	}
	line := lines[0]
	groups := build.FindRegexpGroups(lines[0], ocVersionRe)
	if len(groups) <= 1 {
		return fmt.Errorf(
			"The 'oc' version line '%s' doesn't match the expected regular expression",
			line,
		)
	}
	major, _ := strconv.Atoi(groups["major"])
	minor, _ := strconv.Atoi(groups["minor"])
	micro, _ := strconv.Atoi(groups["micro"])
	if major < minOcMajor || (major == minOcMajor && minor < minOcMinor) {

		return fmt.Errorf(
			"Version %d.%d.%d of 'oc' isn't supported, should be at least %d.%d",
			major, minor, micro,
			minOcMajor, minOcMinor,
		)
	}

	return nil
}

func runOc(args ...string) error {
	err := build.RunCommand("oc", args...)
	if err != nil {
		return fmt.Errorf("The 'oc' command failed: %s", err)
	}
	return nil
}

func evalOc(args ...string) (result string, err error) {
	bytes := build.EvalCommand("oc", args...)
	if bytes == nil {
		err = fmt.Errorf("The 'oc' command failed")
	}
	result = string(bytes)
	return
}
