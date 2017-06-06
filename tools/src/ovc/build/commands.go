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

// This file contains functions intended to simplify the execution and
// evaluation of external commands.

import (
	"os/exec"
	"strings"

	"ovc/log"
)

// RunCommand executes the given command and waits till it finishes. The
// standard output and standard error of the command are redirected to
// the standard output and standard error of the calling program.
//
func RunCommand(name string, args ...string) error {
	log.Debug("Running command '%s' with arguments '%s'", name, strings.Join(args, " "))
	command := exec.Command(name, args...)
	command.Stdout = log.DebugWriter()
	command.Stderr = log.ErrorWriter()
	return command.Run()
}

// EvalCommand executes the given command, waits till it finishes and
// returns the text that it writes to the standard output. If the
// execution of the command fails it returns nil.
//
func EvalCommand(name string, args ...string) []byte {
	log.Debug("Evaluating command '%s' with arguments '%s'", name, strings.Join(args, " "))
	bytes, err := exec.Command(name, args...).Output()
	if err != nil {
		return nil
	}
	return bytes
}
