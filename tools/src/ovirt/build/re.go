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

// This file contains functions useful when working with regular expressions.

package build

import (
	"regexp"
)

// FindRegexGroups checks if line matches the re regular expression. If
// it does match, then it returns a map containing the names of the
// groups as the keys and the text that matches each group as the
// values. If it doesn't match, then it returns nil.
//
func FindRegexpGroups(line string, re *regexp.Regexp) map[string]string {
	groups := make(map[string]string)
	matches := re.FindStringSubmatch(line)
	if matches != nil {
		names := re.SubexpNames()
		for index, name := range names {
			groups[name] = matches[index]
		}
	}
	return groups
}
