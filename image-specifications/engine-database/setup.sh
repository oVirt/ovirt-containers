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

# The base image creates the 'postgresql.conf' file from this template
# each time the container is started. See here for details:
#
#   https://github.com/sclorg/postgresql-container/blob/master/9.5/root/usr/share/container-scripts/postgresql/common.sh
#
# That means that we can modify it, to add our own settings. We check
# that the file exists just fail the build and hence notice early if
# this changes.
POSTGRESQL_CONF_TEMPLATE="${CONTAINER_SCRIPTS_PATH}/openshift-custom-postgresql.conf.template"
if [ ! -f "${POSTGRESQL_CONF_TEMPLATE}" ]; then
    echo "The file '${POSTGRESQL_CONF_TEMPLATE}' doesn't exist." >&2
    exit 1
fi

# Recent versions of the engine need custom auto-vacuum parameters. The
# setup tool checks them, and fails the start of the engine if they
# aren't set.
cat >> "${POSTGRESQL_CONF_TEMPLATE}" <<.
autovacuum_vacuum_scale_factor = 0.01
autovacuum_analyze_scale_factor = 0.075
autovacuum_max_workers = 6
maintenance_work_mem = 65536
.
