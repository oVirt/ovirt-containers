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

FROM centos/postgresql-95-centos7

# The base image changes to user 26 (postgres) in order to run the
# database server. But we need to run the setup as root, as otherwise it
# can't write some of the files. Once that is done we restore user 26.
USER root
COPY setup.sh /
RUN /setup.sh && rm /setup.sh
USER 26
