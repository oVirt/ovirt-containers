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

# Pushing the images to the registry is currently disabled because
# Jenkins doesn't have yet the required credentials.
#make push

clean_up
