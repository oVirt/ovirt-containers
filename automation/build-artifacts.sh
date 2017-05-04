#!/bin/bash -xe
[[ -d exported-artifacts ]] \
|| mkdir -p exported-artifacts

function clean_up {
    make clean
}

trap clean_up SIGHUP SIGINT SIGTERM

# TODO: add jenkins docker login

make build
make push

clean_up
