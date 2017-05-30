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

	"ovirt/build"
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

func main() {
	// Load the project:
	project, err := build.LoadProject("")
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Can't load project: %s\n",
			err,
		)
	}
	defer project.Close()

	// Check that the 'oc' tool is available and that it is the
	// right version:
	validateOc()

	// Log in as system administrator:
	runOc(
		"login",
		"-u",
		"system:admin",
	)

	// Create the project:
	runOc(
		"new-project",
		projectName,
		"--description="+projectTitle,
		"--display-name="+projectTitle,
	)

	// Add administrator permissios to the 'developer' user account:
	runOc(
		"adm",
		"policy",
		"add-role-to-user",
		"admin",
		"developer",
		"-n",
		projectName,
	)

	// Create a service account that can use the root user:
	runOc(
		"create",
		"serviceaccount",
		useRoot,
	)
	runOc(
		"adm",
		"policy",
		"add-scc-to-user",
		"anyuid",
		"-z",
		useRoot,
	)

	// Create a service account that can has access to advanced host
	// privileges, for use inside the VSDC pod:
	runOc(
		"create",
		"serviceaccount",
		privilegedUser,
	)
	runOc(
		"adm",
		"policy",
		"add-scc-to-user",
		"privileged",
		"-z",
		privilegedUser,
	)

	// Create engine and VDSC deployments and add them to the
	// project:
	runOc(
		"create",
		"-f",
		project.Manifests().Directory(),
		"-R",
	)

	// Change the host name for the engine deployment, according to
	// the host name that was assigned to the associated route, then
	// unpause it:
	engineHost := evalOc(
		"get",
		"routes",
		"ovirt-engine",
		"--output=jsonpath={.spec.host}",
	)
	spiceProxyHost := evalOc(
		"get",
		"routes",
		"ovirt-spice-proxy",
		"--output=jsonpath={.spec.host}",
	)
	runOc(
		"set",
		"env",
		"dc/ovirt-engine",
		"-c",
		"ovirt-engine",
		"OVIRT_FQDN="+engineHost,
		"SPICE_PROXY=http://"+spiceProxyHost+":3128",
	)
	runOc(
		"patch",
		"dc/ovirt-engine",
		"--patch",
		`{
			"spec": {
				"paused": false
			}
		}`,
	)
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
func validateOc() {
	// Get the value of the PATH environment variable:
	path, present := os.LookupEnv("PATH")
	if !present {
		fmt.Fprintf(
			os.Stderr,
			"The PATH environment variable isn't set, can't locate the 'oc' tool'.\n",
		)
		os.Exit(1)
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
		fmt.Fprintf(
			os.Stderr,
			"Can't find the 'oc' tool in the path.\n",
		)
		os.Exit(1)
	}

	// Run the tool to extract the version number, and check that it
	// is what we expect:
	out := build.EvalCommand("oc", "version")
	if out == nil {
		fmt.Fprintf(
			os.Stderr,
			"Failed to run 'oc version'.\n",
		)
		os.Exit(1)
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) < 1 {
		fmt.Fprint(
			os.Stderr,
			"Output of 'oc version' is empty.\n",
		)
		os.Exit(1)
	}
	line := lines[0]
	groups := build.FindRegexpGroups(lines[0], ocVersionRe)
	if len(groups) <= 1 {
		fmt.Fprintf(
			os.Stderr,
			"The 'oc' version line '%s' doesn't match the expected regular expression.\n",
			line,
		)
		os.Exit(1)
	}
	major, _ := strconv.Atoi(groups["major"])
	minor, _ := strconv.Atoi(groups["minor"])
	micro, _ := strconv.Atoi(groups["micro"])
	if major < minOcMajor || (major == minOcMajor && minor < minOcMinor) {
		fmt.Fprintf(
			os.Stderr,
			"Version %d.%d.%d of 'oc' isn't supported, should be at least %d.%d.\n",
			major, minor, micro,
			minOcMajor, minOcMinor,
		)
		os.Exit(1)
	}
}

func runOc(args ...string) {
	err := build.RunCommand("oc", args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "The 'oc' command failed: %s\n", err)
		os.Exit(1)
	}
}

func evalOc(args ...string) string {
	result := build.EvalCommand("oc", args...)
	if result == nil {
		fmt.Fprintf(os.Stderr, "The 'oc' command failed\n")
		os.Exit(1)
	}
	return string(result)
}
