#!/bin/bash -xe
[[ -d exported-artifacts ]] \
|| mkdir -p exported-artifacts

function clean_up {
    # We want to return the status of the last command executed *before*
    # cleaning, so we need to save it.
    local status="$?"

    # Do not exit inmediately if the cleaning fails, as we want to
    # report the status of the build, not of the cleaning.
    make clean || true

    exit "${status}"
}

trap clean_up EXIT SIGHUP SIGINT SIGTERM

# TODO: add jenkins docker login

# Build the images:
make build

# Save the images to tar files:
make save

# Move the generated artifacts and log files to the artifacts directory:
mv \
    tools/bin/ovc \
    *.log \
    *.tar.gz \
    exported-artifacts

# Pushing the images to the registry is currently disabled because
# Jenkins doesn't have yet the required credentials.
#make push

clean_up
