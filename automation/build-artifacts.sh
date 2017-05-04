#!/bin/bash -xe
[[ -d exported-artifacts ]] \
|| mkdir -p exported-artifacts

function clean_up {
    make clean
}

trap clean_up SIGHUP SIGINT SIGTERM

# TODO: add jenkins docker login

# Build the images:
make build

# Save the images to tar files, and move them to the exported artifacts
# directory:
make save
mv *.tar.gz exported-artifacts

# Pushing the images to the registry is currently disabled because
# Jenkins doesn't have yet the required credentials.
#make push

clean_up
