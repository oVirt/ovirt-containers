#!/bin/bash -ex

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

# This is script is intended to run during the build of the image, to
# perform all the setup steps generating only one layer.

# Install the oVirt release RPM:
yum -y install http://resources.ovirt.org/pub/yum-repo/ovirt-release41.rpm

# Add here the lines to install other packages that are used/useful in
# all the oVirt containers:
#yum -y install ...
yum -y install centos-release-scl

# Clean the yum database:
yum -y clean all
