#!/bin/bash

mk () {
    pushd $1 >/dev/null && echo $1
    gomake $2
    popd >/dev/null
}

for target in pkg/datastore cmd/shock-server; do
    mk $target $1
done
