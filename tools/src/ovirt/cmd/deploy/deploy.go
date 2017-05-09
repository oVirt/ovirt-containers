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
		"os-manifests",
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
