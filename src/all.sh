#!/bin/bash

mk () {
    pushd $1 >/dev/null && echo $1
    gomake $2
    popd >/dev/null
}

for target in pkg/datastore cmd/shock-server; do
    mk $target $1
done

#pkg/index pkg/index/size pkg/index/fasta pkg/index/fastq pkg/indexer
