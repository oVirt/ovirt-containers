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

# The root of the Go source code:
ROOT=$(PWD)/tools

# Locations of the Go and Glide binaries:
GO_BINARY=go
GLIDE_BINARY=$(ROOT)/bin/glide

# Location of the Glide project:
GLIDE_PROJECT=$(shell find tools/src -name glide.yaml -print -quit)

# Location of the generated tool:
TOOL_BINARY=$(ROOT)/bin/ovc

# Install Glide if necessary:
$(GLIDE_BINARY):
	mkdir -p `dirname $(GLIDE_BINARY)`
	GOPATH="$(ROOT)"; \
	export GOPATH; \
	PATH="$(ROOT)/bin:$${PATH}"; \
	export PATH; \
	curl https://glide.sh/get | sh

# Rule to build the tool from its source code:
$(TOOL_BINARY): $(GLIDE_BINARY) $(shell find tools/src -type f)
	GOPATH="$(ROOT)"; \
	export GOPATH; \
	pushd $$(dirname $(GLIDE_PROJECT)); \
		$(GLIDE_BINARY) install && \
		$(GO_BINARY) build -o $@ *.go || \
		exit 1; \
	popd \

.PHONY: tool
tool: $(TOOL_BINARY)

.PHONY: build
build: $(TOOL_BINARY)
	$< $@ 2>&1 | tee $@.log

.PHONY: save
save: $(TOOL_BINARY)
	$< $@ 2>&1 | tee $@.log

.PHONY: push
push: $(TOOL_BINARY)
	$< $@ 2>&1 | tee $@.log

.PHONY: deploy
deploy: $(TOOL_BINARY)
	$< $@ 2>&1 | tee $@.log

.PHONY: deploy
clean: $(TOOL_BINARY)
	$< $@ 2>&1 | tee $@.log
	rm -rf tools/{bin,pkg}
	docker images --filter dangling=true --quiet | xargs -r docker rmi --force
