#
# Copyright (c) 2017 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# This Makefile is used to build and execute the tools, in particular
# the 'build' tool which is responsible for reading the source code and
# determining what needs to be done to build the images.
#
# If you are looking at how to build the images, then just run 'make',
# and it should take care of it.
#
# If you are looking at changing the build process, then this isn't
# probably the right place. Look at 'tools/src/ovirt/cmd/build/build.go'
# instead, as that is the starting point for the main build process.

.PHONY: all clean

# Go path:
GOPATH="$(PWD)/tools"

# Go dependencies:
GODEPS=\
	gopkg.in/ini.v1 \
	$(NULL)

# Rule to build a tool from its source code:
tools/bin/%: $(shell find tools/src -type f)
	for godep in $(GODEPS); do \
		GOPATH="$(GOPATH)" go get $${godep}; \
	done
	GOPATH="$(GOPATH)" go install ovirt/cmd/$(notdir $@)

build: tools/bin/build
	$< 2>&1 | tee $@.log

push: tools/bin/push
	$< 2>&1 | tee $@.log

clean:
	rm -rf tools/{bin,pkg}
	docker images --filter dangling=true --quiet | xargs -r docker rmi --force
